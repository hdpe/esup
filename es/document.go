package es

import "github.com/tidwall/gjson"

func newDocument(doc gjson.Result) Document {
	return Document{
		id: doc.Get("_id").String(),
		version: Version{
			seqNo:       int(doc.Get("_seq_no").Int()),
			primaryTerm: int(doc.Get("_primary_term").Int()),
		},
		source:    doc.Get("_source"),
		isPresent: true,
	}
}

type Document struct {
	isPresent bool
	id        string
	version   Version
	source    gjson.Result
}

type Version struct {
	seqNo       int
	primaryTerm int
}
