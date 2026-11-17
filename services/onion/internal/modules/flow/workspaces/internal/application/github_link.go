package application

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// githubLinkStatePrefix keeps link and install states apart: each verifier
// rejects the other's shape, and the prefix is inside the signed payload.
const githubLinkStatePrefix = "link"

// GitHubOAuth is the GitHub App's user authorization (web application flow)
// as the application needs it. The user token returned by ExchangeCode is used
// once to read the account and then dropped.
type GitHubOAuth interface {
	AuthorizeURL(state, codeChallenge string) string
	ExchangeCode(ctx context.Context, code, codeVerifier string) (string, error)
	GetUser(ctx context.Context, userToken string) (domain.GitHubUser, error)
}

// SetGitHubOAuth wires member account linking. It needs the GitHub App too,
// whose key signs the link state. nil turns linking off.
func (s *Service) SetGitHubOAuth(oauth GitHubOAuth) { s.github.oauth = oauth }

func (s *Service) githubLinkingConfigured() bool {
	return s.githubConfigured() && s.github.oauth != nil
}

// GitHubUserLink returns the member's linked account in the business, or
// domain.ErrGitHubNotLinked.
func (s *Service) GitHubUserLink(ctx context.Context, businessID, userID int64) (domain.GitHubUserLink, error) {
	return s.repo.GetGitHubUserLink(ctx, businessID, userID)
}

// StartGitHubLink returns GitHub's authorize URL for the member. Linking is
// only offered once the business has connected GitHub.
func (s *Service) StartGitHubLink(ctx context.Context, businessID, userID int64) (string, error) {
	if !s.githubLinkingConfigured() {
		return "", domain.ErrGitHubLinkNotConfigured
	}
	if _, err := s.repo.GetGitHubInstallation(ctx, businessID); err != nil {
		return "", err
	}
	state, err := s.signGitHubLinkState(businessID, userID)
	if err != nil {
		return "", err
	}
	return s.github.oauth.AuthorizeURL(state, githubCodeChallenge(s.githubCodeVerifier(state))), nil
}

// CompleteGitHubLink exchanges the callback code and stores the GitHub
// account for the member who started the link. The state must be the one this
// business and member were given; the code verifier is derived from it, so a
// code obtained for another state cannot be exchanged here.
func (s *Service) CompleteGitHubLink(ctx context.Context, businessID, userID int64, code, state string) (domain.GitHubUserLink, error) {
	if !s.githubLinkingConfigured() {
		return domain.GitHubUserLink{}, domain.ErrGitHubLinkNotConfigured
	}
	if err := s.verifyGitHubLinkState(state, businessID, userID); err != nil {
		return domain.GitHubUserLink{}, err
	}
	if _, err := s.repo.GetGitHubInstallation(ctx, businessID); err != nil {
		return domain.GitHubUserLink{}, err
	}
	token, err := s.github.oauth.ExchangeCode(ctx, code, s.githubCodeVerifier(state))
	if err != nil {
		return domain.GitHubUserLink{}, err
	}
	user, err := s.github.oauth.GetUser(ctx, token)
	if err != nil {
		return domain.GitHubUserLink{}, err
	}
	err = s.repo.ReplaceGitHubUserLink(ctx, domain.GitHubUserLink{
		BusinessID:   businessID,
		UserID:       userID,
		GitHubUserID: user.ID,
		Login:        user.Login,
		AvatarURL:    user.AvatarURL,
	})
	if err != nil {
		return domain.GitHubUserLink{}, err
	}
	return s.repo.GetGitHubUserLink(ctx, businessID, userID)
}

// UnlinkGitHub removes the member's linked account. Unlinking when nothing is
// linked succeeds.
func (s *Service) UnlinkGitHub(ctx context.Context, businessID, userID int64) error {
	return s.repo.DeleteGitHubUserLink(ctx, businessID, userID)
}

// Link state: link.business.user.issuedUnix.nonce.signature, HMAC-SHA256 over
// everything before the last dot, with the install state's key.
func (s *Service) signGitHubLinkState(businessID, userID int64) (string, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := fmt.Sprintf("%s.%d.%d.%d.%s", githubLinkStatePrefix, businessID, userID, s.github.now().Unix(), hex.EncodeToString(nonce))
	return payload + "." + s.githubStateMAC(payload), nil
}

func (s *Service) verifyGitHubLinkState(state string, businessID, userID int64) error {
	cut := strings.LastIndex(state, ".")
	if cut <= 0 {
		return domain.ErrGitHubInvalidState
	}
	payload, sig := state[:cut], state[cut+1:]
	if !hmac.Equal([]byte(sig), []byte(s.githubStateMAC(payload))) {
		return domain.ErrGitHubInvalidState
	}
	parts := strings.Split(payload, ".")
	if len(parts) != 5 || parts[0] != githubLinkStatePrefix {
		return domain.ErrGitHubInvalidState
	}
	stateBusiness, errB := strconv.ParseInt(parts[1], 10, 64)
	stateUser, errU := strconv.ParseInt(parts[2], 10, 64)
	issuedUnix, errT := strconv.ParseInt(parts[3], 10, 64)
	if errB != nil || errU != nil || errT != nil || stateBusiness != businessID || stateUser != userID {
		return domain.ErrGitHubInvalidState
	}
	issued := time.Unix(issuedUnix, 0)
	now := s.github.now()
	if now.Sub(issued) > githubStateTTL || issued.After(now.Add(time.Minute)) {
		return domain.ErrGitHubInvalidState
	}
	return nil
}

// githubCodeVerifier derives the PKCE verifier from the signed state, so
// nothing is stored between start and complete. Only the server knows the
// key, so the verifier cannot be computed from the state alone. 32 bytes
// base64url-encode to the 43 characters PKCE requires at minimum.
func (s *Service) githubCodeVerifier(state string) string {
	mac := hmac.New(sha256.New, s.github.stateKey)
	mac.Write([]byte("pkce\x00" + state))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func githubCodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
