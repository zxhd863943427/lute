// Lute - A structured markdown engine.
// Copyright (C) 2019-present, b3log.org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lute

type Blockquote struct {
	*BaseNode
}

func (t *Tree) parseBlockquote(line items)  {
	_, line = line.trimLeft()
	blockquote := &Blockquote{&BaseNode{typ: NodeBlockquote}}
	t.context.BlockContainers.push(blockquote)
	line = line[1:]
	if line[0].isSpace() {
		line = line[1:]
	} else if line[0].isTab() {
		line = t.indentOffset(line, 2)
	}

	for {
		t.parseBlock(line)

		_, line = line.trimLeft()
		line = t.trimBlockquoteMarker(line)

		break
	}

	t.context.BlockContainers.pop()
}

func (t *Tree) isBlockquote(line items) bool {
	if 2 > len(line) { // at least > and newline
		return false
	}

	_, marker := line.firstNonSpace()
	if itemGreater != marker.typ {
		return false
	}

	return true
}

func (t *Tree) trimBlockquoteMarker(line items) (ret items) {
	if NodeBlockquote != t.context.BlockContainers.peek().Type() {
		return line
	}

	count := 1
	i := 0
	for _, ret = line[1:].trimLeft(); 0 < len(ret) && (itemGreater == ret[0].typ || ret[0].isSpaceOrTab()); ret = ret[1:] {
		if i++; count < i {
			break
		}
	}

	return
}

func (t *Tree) isParagraphContinuation(line items) bool {
	lastc := t.context.BlockContainers.peek().LastChild()
	if nil == lastc {
		return false
	}

	if NodeParagraph != lastc.Type() {
		return false
	}

	line = t.trimBlockquoteMarker(line)
	startIndentSpaces := line.spaceCountLeft()
	if t.interruptParagraph(startIndentSpaces, line) {
		return false
	}

	if line.isBlankLine() {
		return false
	}

	return true
}

func (t *Tree) blockquoteMarkerCount(line items) (ret int) {
	_, line = line.trimLeft()
	for _, token := range line {
		if itemGreater == token.typ {
			ret++
		} else if itemSpace != token.typ && itemTab != token.typ {
			break
		}
	}

	return
}

func (t *Tree) skipBlankBlockquote() (line items) {
	for {
		line = t.nextLine()
		if line.isEOF() {
			return
		}

		if !t.isBlockquote(line) {
			return line
		}

		remains := t.trimBlockquoteMarker(line)
		if remains.isBlankLine() {
			continue
		}

		return remains
	}
}
