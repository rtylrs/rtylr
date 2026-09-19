package model

import "fmt"

// Example is one scoring decision: a form element's context plus the
// candidate values extracted from the source document.
type Example struct {
	Context string
	Options []string
}

func byteIDs(s string, length int) []int {
	raw := []byte(s)
	if len(raw) > length {
		raw = raw[:length]
	}
	ids := make([]int, len(raw))
	for i, b := range raw {
		ids[i] = int(b) + 1
	}
	return ids
}

type batch struct {
	ctxIDs  [][]int
	ctxMask [][]bool
	optIDs  [][][]int
	optMask [][]bool
	optTok  [][][]bool
}

func collate(exs []Example, ctxLen, optLen int) (batch, error) {
	var b batch
	maxCtx, maxOpts, maxTok := 1, 0, 1
	rows := make([]struct {
		ctx  []int
		opts [][]int
	}, len(exs))
	for i, ex := range exs {
		if ex.Context == "" {
			return b, fmt.Errorf("s1: example %d has an empty context", i)
		}
		if len(ex.Options) == 0 {
			return b, fmt.Errorf("s1: example %d has no options", i)
		}
		rows[i].ctx = byteIDs(ex.Context, ctxLen)
		for j, opt := range ex.Options {
			if opt == "" {
				return b, fmt.Errorf("s1: example %d option %d is empty", i, j)
			}
			rows[i].opts = append(rows[i].opts, byteIDs(opt, optLen))
		}
		if len(rows[i].ctx) > maxCtx {
			maxCtx = len(rows[i].ctx)
		}
		if len(rows[i].opts) > maxOpts {
			maxOpts = len(rows[i].opts)
		}
		for _, ids := range rows[i].opts {
			if len(ids) > maxTok {
				maxTok = len(ids)
			}
		}
	}
	b.ctxIDs = make([][]int, len(exs))
	b.ctxMask = make([][]bool, len(exs))
	b.optIDs = make([][][]int, len(exs))
	b.optMask = make([][]bool, len(exs))
	b.optTok = make([][][]bool, len(exs))
	for i, row := range rows {
		b.ctxIDs[i] = make([]int, maxCtx)
		b.ctxMask[i] = make([]bool, maxCtx)
		copy(b.ctxIDs[i], row.ctx)
		for t := range row.ctx {
			b.ctxMask[i][t] = true
		}
		b.optIDs[i] = make([][]int, maxOpts)
		b.optMask[i] = make([]bool, maxOpts)
		b.optTok[i] = make([][]bool, maxOpts)
		for o, ids := range row.opts {
			b.optIDs[i][o] = make([]int, maxTok)
			b.optTok[i][o] = make([]bool, maxTok)
			copy(b.optIDs[i][o], ids)
			b.optMask[i][o] = true
			for t := range ids {
				b.optTok[i][o][t] = true
			}
		}
	}
	return b, nil
}
