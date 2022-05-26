package gin

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/genproto/googleapis/api/annotations"
	"io/ioutil"
	"strings"
)

type httpRule struct {
	method  string
	pattern string
	body    string
}

type protoExtend struct {
	protoName string
	methodMap map[string]*annotations.HttpRule
}

func newProtoExtend(protoName string) (pe *protoExtend, err error) {
	pe = &protoExtend{
		protoName: protoName,
		methodMap: make(map[string]*annotations.HttpRule),
	}

	if protoName == "" {
		return
	}

	err = pe.setup()
	return
}

func (p *protoExtend) setup() (err error) {
	desc, err := decodeFileDesc(proto.FileDescriptor(p.protoName))
	if err != nil {
		return
	}

	for _, service := range desc.Service {
		for _, md := range service.Method {
			if md == nil || md.Options == nil {
				continue
			}

			extend, err := proto.GetExtension(md.Options, annotations.E_Http)
			if err != nil {
				return err
			}

			hr, ok := extend.(*annotations.HttpRule)
			if !ok {
				return errors.New("option invalid")
			}
			p.methodMap[strings.Title(*md.Name)] = hr
		}
	}
	return
}

// 可能匹配多个 rest 接口
func (p *protoExtend) methodHttpRules(method string) (rules []httpRule) {
	ruler, ok := p.methodMap[method]
	if !ok {
		rules = append(rules, httpRule{method: "POST", pattern: ToUnderLine(method), body: "*"})
		return
	}

	md, pattern := methodPattern(ruler)
	rules = append(rules, httpRule{method: md, pattern: pattern, body: ruler.Body})
	for _, r := range ruler.AdditionalBindings {
		md, pattern := methodPattern(r)
		rules = append(rules, httpRule{method: md, pattern: pattern, body: ruler.Body})
	}
	return
}

func methodPattern(hr *annotations.HttpRule) (method, pattern string) {
	if get, ok := hr.Pattern.(*annotations.HttpRule_Get); ok {
		return "GET", ginMethodPattern(get.Get)
	} else if post, ok := hr.Pattern.(*annotations.HttpRule_Post); ok {
		return "POST", ginMethodPattern(post.Post)
	} else if patch, ok := hr.Pattern.(*annotations.HttpRule_Patch); ok {
		return "PATCH", ginMethodPattern(patch.Patch)
	} else if put, ok := hr.Pattern.(*annotations.HttpRule_Put); ok {
		return "PUT", ginMethodPattern(put.Put)
	} else if del, ok := hr.Pattern.(*annotations.HttpRule_Delete); ok {
		return "DELETE", ginMethodPattern(del.Delete)
	} else if cus, ok := hr.Pattern.(*annotations.HttpRule_Custom); ok {
		return cus.Custom.Kind, ginMethodPattern(cus.Custom.Path)
	}
	panic(fmt.Sprintf("can not support ruler:%v", hr.Pattern))
}

func ginMethodPattern(pattern string) string {
	pattern = strings.Replace(pattern, "{", ":", -1)
	return strings.Replace(pattern, "}", "", -1)
}

func decodeFileDesc(enc []byte) (*dpb.FileDescriptorProto, error) {
	raw, err := decompress(enc)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress enc: %v", err)
	}

	fd := new(dpb.FileDescriptorProto)
	if err := proto.Unmarshal(raw, fd); err != nil {
		return nil, fmt.Errorf("bad descriptor: %v", err)
	}
	return fd, nil
}

// decompress does gzip decompression.
func decompress(b []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("bad gzipped descriptor: %v", err)
	}
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("bad gzipped descriptor: %v", err)
	}
	return out, nil
}
