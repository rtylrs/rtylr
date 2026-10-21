package application

// submissionNotice marks a public submission that needs a human to step in.
// It changes the subject and adds a banner to the team notification email, so
// a degraded submission is never indistinguishable from a normal one.
type submissionNotice int

const (
	noticeNone submissionNotice = iota
	// noticeCVUploadFailed: the application row was saved, but the attached CV
	// could not be stored. Someone has to ask the applicant for it.
	noticeCVUploadFailed
	// noticeNotSaved: the database rejected the application. This email is the
	// only remaining record of the applicant.
	noticeNotSaved
)

func noticeForSaved(cvUploadFailed bool) submissionNotice {
	if cvUploadFailed {
		return noticeCVUploadFailed
	}
	return noticeNone
}

func (n submissionNotice) subjectPrefix() string {
	switch n {
	case noticeCVUploadFailed:
		return "[ACTION NEEDED: CV not stored] "
	case noticeNotSaved:
		return "[ACTION NEEDED: application NOT saved] "
	default:
		return ""
	}
}

func (n submissionNotice) bannerHTML() string {
	const open = `<div style="margin:0 0 24px 0;padding:14px 16px;background:rgba(235,87,87,0.15);border-radius:8px;border-left:3px solid #eb5757;font-size:14px;line-height:1.6;color:#f0f0f6;">`
	switch n {
	case noticeCVUploadFailed:
		return open + `<strong>The CV could not be stored.</strong> The application is saved without it. Ask the applicant to send their CV by email or WhatsApp.</div>`
	case noticeNotSaved:
		return open + `<strong>This application was NOT saved to the database.</strong> This email is the only record. The applicant was shown an error. Contact them and re-enter it by hand.</div>`
	default:
		return ""
	}
}
