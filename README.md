## Validator

[![GoDoc Widget](https://godoc.org/github.com/go-courier/validator?status.svg)](https://godoc.org/github.com/go-courier/validator)
[![Build Status](https://travis-ci.org/go-courier/validator.svg?branch=master)](https://travis-ci.org/go-courier/validator)
[![codecov](https://codecov.io/gh/go-courier/validator/branch/master/graph/badge.svg)](https://codecov.io/gh/go-courier/validator)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-courier/validator)](https://goreportcard.com/report/github.com/go-courier/validator)


## Rules

```
// simple 
@name

// with parameters 
@name<param1>
@name<param1,param2>

// with ranges
@name[from, to)
@name[length]

// with values 
@name{VALUE1,VALUE2,VALUE3}
@name{%v}

// with regexp
@name/\d+/

// composes
@map<@string[1,10],@string{A,B,C}>
@map<@string[1,10],@string/\d+/>[0,10]
```


## Built-in rules

* [@string](https://godoc.org/github.com/go-courier/validator#StringValidator)
* [@int](https://godoc.org/github.com/go-courier/validator#IntValidator)
* [@uint](https://godoc.org/github.com/go-courier/validator#UintValidator)
* [@float](https://godoc.org/github.com/go-courier/validator#FloatValidator)
* [@map](https://godoc.org/github.com/go-courier/validator#MapValidator)
* [@slice](https://godoc.org/github.com/go-courier/validator#SliceValidator)

