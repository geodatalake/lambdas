// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"

func Handle(evt interface{}, ctx *runtime.Context) (string, error) {
	return "hello World!", nil
}
