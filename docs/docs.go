// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2018-11-04 21:12:10.529569 +0800 CST m=+0.079119212

package docs

import (
	"github.com/swaggo/swag"
)

var doc = `{
    "swagger": "2.0",
    "info": {
        "description": "A read/write API for the Matterbridge chat bridge.",
        "title": "Matterbridge API",
        "contact": {},
        "license": {
            "name": "Apache 2.0",
            "url": "https://github.com/42wim/matterbridge/blob/master/LICENSE"
        },
        "version": "TODO"
    },
    "host": "TODO",
    "basePath": "/api",
    "paths": {}
}`

type s struct{}

func (s *s) ReadDoc() string {
	return doc
}
func init() {
	swag.Register(swag.Name, &s{})
}
