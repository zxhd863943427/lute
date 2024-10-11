// Lute - 一款结构化的 Markdown 引擎，支持 Go 和 JavaScript
// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package render

import (
	"errors"
	"strings"

	"github.com/dgraph-io/ristretto"
	"github.com/goccy/go-json"

	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
	"github.com/88250/lute/util"
)

var JSONcache, _ = ristretto.NewCache(&ristretto.Config[string, string]{
	NumCounters: 1024 * 1024 * 200,
	MaxCost:     1024 * 1024 * 200,
	BufferItems: 64,
})

func putJSON(path, json string) {
	JSONcache.Set(path, json, 16)
}

func getJSON(path string) (ret string, found bool) {
	ret, found = JSONcache.Get(path)
	return
}

func getNodePath(node *ast.Node) (ret string, err error) {
	var builder strings.Builder
	if node.ID != "" {
		builder.WriteString(node.ID)
		builder.WriteString(node.IALAttr("updated"))
		ret = builder.String()
		if len(ret) < 1 {
			err = errors.New("path length = 0")
		}
		return
	}
	if node.Parent != nil {
		index := 0
		parent := node.Parent
		for c := node.Previous; c != nil; c = c.Previous {
			index++
		}
		builder.WriteString(parent.ID)
		builder.WriteString(parent.IALAttr("updated"))
		builder.WriteString(":")
		builder.WriteRune(rune(index))
		ret = builder.String()

		if len(ret) < 1 {
			err = errors.New("path length = 0")
		}
		return
	}
	err = errors.New("gen node path error")
	return
}

type JSONRenderer struct {
	*BaseRenderer
}

func NewJSONRenderer(tree *parse.Tree, options *Options) Renderer {
	var ials []*ast.Node // 渲染器剔除语法树块级 IAL 节点
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if ast.NodeKramdownBlockIAL == n.Type {
			ials = append(ials, n)
		}
		return ast.WalkContinue
	})
	for _, ial := range ials {
		ial.Unlink()
	}

	ret := &JSONRenderer{NewBaseRenderer(tree, options)}
	ret.DefaultRendererFunc = ret.renderNode
	return ret
}

func tryGetCache(node *ast.Node) (ret string, err error) {
	path, err := getNodePath(node)
	if err != nil {
		return
	}
	cacheData, found := getJSON(path)
	if !found {
		err = errors.New("no cache")
		return
	}
	ret = cacheData
	return
}

func (r *JSONRenderer) renderNode(node *ast.Node, entering bool) ast.WalkStatus {
	if entering {
		if nil != node.Previous {
			r.WriteString(",")
		}
		var n string

		cacheData, err := tryGetCache(node)
		if err == nil {
			n = cacheData
		} else {
			n, err = buildJSON(node)
			if nil != err {
				panic("marshal node to json failed: " + err.Error())
				return ast.WalkStop
			}
			path, err := getNodePath(node)
			if err == nil {
				putJSON(path, n)
			}

		}
		n = n[:len(n)-1] // 去掉结尾的 }
		r.WriteString(n)
		if nil != node.FirstChild {
			r.WriteString(",\"Children\":[")
		} else {
			r.WriteString("}")
		}
	} else {
		if nil != node.FirstChild {
			r.WriteByte(']')
			r.WriteString("}")
		}
	}
	return ast.WalkContinue
}

func buildJSON(n *ast.Node) (jsonData string, err error) {
	n_clone := n.Clone()
	n_clone.Data, n_clone.TypeStr = util.BytesToStr(n_clone.Tokens), n_clone.Type.String()
	n_clone.Properties = ial2Map(n_clone.KramdownIAL)
	delete(n_clone.Properties, "refcount")
	delete(n_clone.Properties, "av-names")
	data, err := json.Marshal(n_clone)
	if nil != err {
		return
	}
	jsonData = util.BytesToStr(data)
	return
}

func ial2Map(ial [][]string) (ret map[string]string) {
	ret = map[string]string{}
	for _, kv := range ial {
		ret[kv[0]] = kv[1]
	}
	return
}
