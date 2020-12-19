package json

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"unsafe"
)

const startDetectingCyclesAfter = 1000

func load(base uintptr, idx uintptr) uintptr {
	addr := base + idx
	return **(**uintptr)(unsafe.Pointer(&addr))
}

func store(base uintptr, idx uintptr, p uintptr) {
	addr := base + idx
	**(**uintptr)(unsafe.Pointer(&addr)) = p
}

func errUnsupportedValue(code *opcode, ptr uintptr) *UnsupportedValueError {
	v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: code.typ,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&ptr)),
	}))
	return &UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   fmt.Sprintf("encountered a cycle via %s", code.typ),
	}
}

func errUnsupportedFloat(v float64) *UnsupportedValueError {
	return &UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   strconv.FormatFloat(v, 'g', -1, 64),
	}
}

func errMarshaler(code *opcode, err error) *MarshalerError {
	return &MarshalerError{
		Type: rtype2type(code.typ),
		Err:  err,
	}
}

func (e *Encoder) run(ctx *encodeRuntimeContext, b []byte, code *opcode) ([]byte, error) {
	recursiveLevel := 0
	seenPtr := map[uintptr]struct{}{}
	ptrOffset := uintptr(0)
	ctxptr := ctx.ptr()

	for {
		switch code.op {
		default:
			return nil, fmt.Errorf("failed to handle opcode. doesn't implement %s", code.op)
		case opPtr, opPtrIndent:
			ptr := load(ctxptr, code.idx)
			code = code.next
			store(ctxptr, code.idx, e.ptrToPtr(ptr))
		case opInt:
			b = encodeInt(b, e.ptrToInt(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opIntIndent:
			b = encodeInt(b, e.ptrToInt(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opInt8:
			b = encodeInt8(b, e.ptrToInt8(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opInt8Indent:
			b = encodeInt8(b, e.ptrToInt8(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opInt16:
			b = encodeInt16(b, e.ptrToInt16(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opInt16Indent:
			b = encodeInt16(b, e.ptrToInt16(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opInt32:
			b = encodeInt32(b, e.ptrToInt32(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opInt32Indent:
			b = encodeInt32(b, e.ptrToInt32(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opInt64:
			b = encodeInt64(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opInt64Indent:
			b = encodeInt64(b, e.ptrToInt64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUint:
			b = encodeUint(b, e.ptrToUint(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opUintIndent:
			b = encodeUint(b, e.ptrToUint(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUint8:
			b = encodeUint8(b, e.ptrToUint8(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opUint8Indent:
			b = encodeUint8(b, e.ptrToUint8(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUint16:
			b = encodeUint16(b, e.ptrToUint16(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opUint16Indent:
			b = encodeUint16(b, e.ptrToUint16(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUint32:
			b = encodeUint32(b, e.ptrToUint32(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opUint32Indent:
			b = encodeUint32(b, e.ptrToUint32(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opUint64:
			b = encodeUint64(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opUint64Indent:
			b = encodeUint64(b, e.ptrToUint64(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opFloat32:
			b = encodeFloat32(b, e.ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opFloat32Indent:
			b = encodeFloat32(b, e.ptrToFloat32(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opFloat64:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opFloat64Indent:
			v := e.ptrToFloat64(load(ctxptr, code.idx))
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opString:
			b = e.encodeString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opStringIndent:
			b = e.encodeString(b, e.ptrToString(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBool:
			b = encodeBool(b, e.ptrToBool(load(ctxptr, code.idx)))
			b = encodeComma(b)
			code = code.next
		case opBoolIndent:
			b = encodeBool(b, e.ptrToBool(load(ctxptr, code.idx)))
			b = encodeIndentComma(b)
			code = code.next
		case opBytes:
			ptr := load(ctxptr, code.idx)
			slice := e.ptrToSlice(ptr)
			if ptr == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, e.ptrToBytes(ptr))
			}
			b = encodeComma(b)
			code = code.next
		case opBytesIndent:
			ptr := load(ctxptr, code.idx)
			slice := e.ptrToSlice(ptr)
			if ptr == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
			} else {
				b = encodeByteSlice(b, e.ptrToBytes(ptr))
			}
			b = encodeIndentComma(b)
			code = code.next
		case opInterface:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			if _, exists := seenPtr[ptr]; exists {
				return nil, errUnsupportedValue(code, ptr)
			}
			seenPtr[ptr] = struct{}{}
			v := e.ptrToInterface(code, ptr)
			ctx.keepRefs = append(ctx.keepRefs, unsafe.Pointer(&v))
			rv := reflect.ValueOf(v)
			if rv.IsNil() {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			vv := rv.Interface()
			header := (*interfaceHeader)(unsafe.Pointer(&vv))
			typ := header.typ
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			var c *opcode
			if typ.Kind() == reflect.Map {
				code, err := e.compileMap(&encodeCompileContext{
					typ:        typ,
					root:       code.root,
					withIndent: e.enabledIndent,
					indent:     code.indent,
				}, false)
				if err != nil {
					return nil, err
				}
				c = code
			} else {
				code, err := e.compile(&encodeCompileContext{
					typ:        typ,
					root:       code.root,
					withIndent: e.enabledIndent,
					indent:     code.indent,
				})
				if err != nil {
					return nil, err
				}
				c = code
			}

			beforeLastCode := c.beforeLastCode()
			lastCode := beforeLastCode.next
			lastCode.idx = beforeLastCode.idx + uintptrSize
			totalLength := uintptr(code.totalLength())
			nextTotalLength := uintptr(c.totalLength())
			curlen := uintptr(len(ctx.ptrs))
			offsetNum := ptrOffset / uintptrSize
			oldOffset := ptrOffset
			ptrOffset += totalLength * uintptrSize

			newLen := offsetNum + totalLength + nextTotalLength
			if curlen < newLen {
				ctx.ptrs = append(ctx.ptrs, make([]uintptr, newLen-curlen)...)
			}
			ctxptr = ctx.ptr() + ptrOffset // assign new ctxptr

			store(ctxptr, 0, uintptr(header.ptr))
			store(ctxptr, lastCode.idx, oldOffset)

			// link lastCode ( opInterfaceEnd ) => code.next
			lastCode.op = opInterfaceEnd
			lastCode.next = code.next

			code = c
			recursiveLevel++
		case opInterfaceIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.next
				break
			}
			if _, exists := seenPtr[ptr]; exists {
				return nil, errUnsupportedValue(code, ptr)
			}
			seenPtr[ptr] = struct{}{}
			v := e.ptrToInterface(code, ptr)
			ctx.keepRefs = append(ctx.keepRefs, unsafe.Pointer(&v))
			rv := reflect.ValueOf(v)
			if rv.IsNil() {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.next
				break
			}
			vv := rv.Interface()
			header := (*interfaceHeader)(unsafe.Pointer(&vv))
			typ := header.typ
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			var c *opcode
			if typ.Kind() == reflect.Map {
				code, err := e.compileMap(&encodeCompileContext{
					typ:        typ,
					root:       code.root,
					withIndent: e.enabledIndent,
					indent:     code.indent,
				}, false)
				if err != nil {
					return nil, err
				}
				c = code
			} else {
				code, err := e.compile(&encodeCompileContext{
					typ:        typ,
					root:       code.root,
					withIndent: e.enabledIndent,
					indent:     code.indent,
				})
				if err != nil {
					return nil, err
				}
				c = code
			}

			beforeLastCode := c.beforeLastCode()
			lastCode := beforeLastCode.next
			lastCode.idx = beforeLastCode.idx + uintptrSize
			totalLength := uintptr(code.totalLength())
			nextTotalLength := uintptr(c.totalLength())
			curlen := uintptr(len(ctx.ptrs))
			offsetNum := ptrOffset / uintptrSize
			oldOffset := ptrOffset
			ptrOffset += totalLength * uintptrSize

			newLen := offsetNum + totalLength + nextTotalLength
			if curlen < newLen {
				ctx.ptrs = append(ctx.ptrs, make([]uintptr, newLen-curlen)...)
			}
			ctxptr = ctx.ptr() + ptrOffset // assign new ctxptr

			store(ctxptr, 0, uintptr(header.ptr))
			store(ctxptr, lastCode.idx, oldOffset)

			// link lastCode ( opInterfaceEnd ) => code.next
			lastCode.op = opInterfaceEnd
			lastCode.next = code.next

			code = c
			recursiveLevel++
		case opInterfaceEnd, opInterfaceEndIndent:
			recursiveLevel--
			// restore ctxptr
			offset := load(ctxptr, code.idx)
			ctxptr = ctx.ptr() + offset
			ptrOffset = offset
			code = code.next
		case opMarshalJSON:
			ptr := load(ctxptr, code.idx)
			v := e.ptrToInterface(code, ptr)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			if len(bb) == 0 {
				return nil, errUnexpectedEndOfJSON(
					fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
					0,
				)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = append(append(b, buf.Bytes()...), ',')
			code = code.next
		case opMarshalJSONIndent:
			ptr := load(ctxptr, code.idx)
			v := e.ptrToInterface(code, ptr)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			if len(bb) == 0 {
				return nil, errUnexpectedEndOfJSON(
					fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
					0,
				)
			}
			var buf bytes.Buffer
			if err := encodeWithIndent(
				&buf,
				bb,
				string(e.prefix)+string(bytes.Repeat(e.indentStr, code.indent)),
				string(e.indentStr),
			); err != nil {
				return nil, err
			}
			b = append(b, buf.Bytes()...)
			b = encodeIndentComma(b)
			code = code.next
		case opMarshalText:
			ptr := load(ctxptr, code.idx)
			isPtr := code.typ.Kind() == reflect.Ptr
			p := e.ptrToUnsafePtr(ptr)
			if p == nil {
				b = encodeNull(b)
				b = encodeComma(b)
			} else if isPtr && *(*unsafe.Pointer)(p) == nil {
				b = append(b, '"', '"', ',')
			} else {
				if isPtr && code.typ.Elem().Implements(marshalTextType) {
					p = *(*unsafe.Pointer)(p)
				}
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: p,
				}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
			}
			code = code.next
		case opMarshalTextIndent:
			ptr := load(ctxptr, code.idx)
			isPtr := code.typ.Kind() == reflect.Ptr
			p := e.ptrToUnsafePtr(ptr)
			if p == nil {
				b = encodeNull(b)
				b = encodeIndentComma(b)
			} else if isPtr && *(*unsafe.Pointer)(p) == nil {
				b = append(b, '"', '"', ',', '\n')
			} else {
				if isPtr && code.typ.Elem().Implements(marshalTextType) {
					p = *(*unsafe.Pointer)(p)
				}
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
					typ: code.typ,
					ptr: p,
				}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeIndentComma(b)
			}
			code = code.next
		case opSliceHead:
			p := load(ctxptr, code.idx)
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				store(ctxptr, code.elemIdx, 0)
				store(ctxptr, code.length, uintptr(slice.len))
				store(ctxptr, code.idx, uintptr(slice.data))
				if slice.len > 0 {
					b = append(b, '[')
					code = code.next
					store(ctxptr, code.idx, uintptr(slice.data))
				} else {
					b = append(b, '[', ']', ',')
					code = code.end.next
				}
			}
		case opSliceElem:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if idx < length {
				store(ctxptr, code.elemIdx, idx)
				data := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, data+idx*size)
			} else {
				last := len(b) - 1
				b[last] = ']'
				b = encodeComma(b)
				code = code.end.next
			}
		case opSliceHeadIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				slice := e.ptrToSlice(p)
				store(ctxptr, code.elemIdx, 0)
				store(ctxptr, code.length, uintptr(slice.len))
				store(ctxptr, code.idx, uintptr(slice.data))
				if slice.len > 0 {
					b = append(b, '[', '\n')
					b = e.encodeIndent(b, code.indent+1)
					code = code.next
					store(ctxptr, code.idx, uintptr(slice.data))
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '[', ']', '\n')
					code = code.end.next
				}
			}
		case opRootSliceHeadIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				slice := e.ptrToSlice(p)
				store(ctxptr, code.elemIdx, 0)
				store(ctxptr, code.length, uintptr(slice.len))
				store(ctxptr, code.idx, uintptr(slice.data))
				if slice.len > 0 {
					b = append(b, '[', '\n')
					b = e.encodeIndent(b, code.indent+1)
					code = code.next
					store(ctxptr, code.idx, uintptr(slice.data))
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '[', ']', ',', '\n')
					code = code.end.next
				}
			}
		case opSliceElemIndent:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if idx < length {
				b = e.encodeIndent(b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				data := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, data+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = e.encodeIndent(b, code.indent)
				b = append(b, ']', ',', '\n')
				code = code.end.next
			}
		case opRootSliceElemIndent:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if idx < length {
				b = e.encodeIndent(b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				code = code.next
				data := load(ctxptr, code.headIdx)
				store(ctxptr, code.idx, data+idx*code.size)
			} else {
				b = append(b, '\n')
				b = e.encodeIndent(b, code.indent)
				b = append(b, ']')
				code = code.end.next
			}
		case opArrayHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				if code.length > 0 {
					b = append(b, '[')
					store(ctxptr, code.elemIdx, 0)
					code = code.next
					store(ctxptr, code.idx, p)
				} else {
					b = append(b, '[', ']', ',')
					code = code.end.next
				}
			}
		case opArrayElem:
			idx := load(ctxptr, code.elemIdx)
			idx++
			if idx < code.length {
				store(ctxptr, code.elemIdx, idx)
				p := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, p+idx*size)
			} else {
				last := len(b) - 1
				b[last] = ']'
				b = encodeComma(b)
				code = code.end.next
			}
		case opArrayHeadIndent:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				if code.length > 0 {
					b = append(b, '[', '\n')
					b = e.encodeIndent(b, code.indent+1)
					store(ctxptr, code.elemIdx, 0)
					code = code.next
					store(ctxptr, code.idx, p)
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '[', ']', ',', '\n')
					code = code.end.next
				}
			}
		case opArrayElemIndent:
			idx := load(ctxptr, code.elemIdx)
			idx++
			if idx < code.length {
				b = e.encodeIndent(b, code.indent+1)
				store(ctxptr, code.elemIdx, idx)
				p := load(ctxptr, code.headIdx)
				size := code.size
				code = code.next
				store(ctxptr, code.idx, p+idx*size)
			} else {
				b = b[:len(b)-2]
				b = append(b, '\n')
				b = e.encodeIndent(b, code.indent)
				b = append(b, ']', ',', '\n')
				code = code.end.next
			}
		case opMapHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				uptr := e.ptrToUnsafePtr(ptr)
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{')
					iter := mapiterinit(code.typ, uptr)
					ctx.keepRefs = append(ctx.keepRefs, iter)
					store(ctxptr, code.elemIdx, 0)
					store(ctxptr, code.length, uintptr(mlen))
					store(ctxptr, code.mapIter, uintptr(iter))
					if !e.unorderedMap {
						pos := make([]int, 0, mlen)
						pos = append(pos, len(b))
						posPtr := unsafe.Pointer(&pos)
						ctx.keepRefs = append(ctx.keepRefs, posPtr)
						store(ctxptr, code.end.mapPos, uintptr(posPtr))
					}
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					b = append(b, '{', '}', ',')
					code = code.end.next
				}
			}
		case opMapHeadLoad:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				// load pointer
				ptr = e.ptrToPtr(ptr)
				uptr := e.ptrToUnsafePtr(ptr)
				if ptr == 0 {
					b = encodeNull(b)
					b = encodeComma(b)
					code = code.end.next
					break
				}
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{')
					iter := mapiterinit(code.typ, uptr)
					ctx.keepRefs = append(ctx.keepRefs, iter)
					store(ctxptr, code.elemIdx, 0)
					store(ctxptr, code.length, uintptr(mlen))
					store(ctxptr, code.mapIter, uintptr(iter))
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					if !e.unorderedMap {
						pos := make([]int, 0, mlen)
						pos = append(pos, len(b))
						posPtr := unsafe.Pointer(&pos)
						ctx.keepRefs = append(ctx.keepRefs, posPtr)
						store(ctxptr, code.end.mapPos, uintptr(posPtr))
					}
					code = code.next
				} else {
					b = append(b, '{', '}', ',')
					code = code.end.next
				}
			}
		case opMapKey:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if e.unorderedMap {
				if idx < length {
					ptr := load(ctxptr, code.mapIter)
					iter := e.ptrToUnsafePtr(ptr)
					store(ctxptr, code.elemIdx, idx)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					last := len(b) - 1
					b[last] = '}'
					b = encodeComma(b)
					code = code.end.next
				}
			} else {
				ptr := load(ctxptr, code.end.mapPos)
				posPtr := (*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
				*posPtr = append(*posPtr, len(b))
				if idx < length {
					ptr := load(ctxptr, code.mapIter)
					iter := e.ptrToUnsafePtr(ptr)
					store(ctxptr, code.elemIdx, idx)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					code = code.end
				}
			}
		case opMapValue:
			if e.unorderedMap {
				last := len(b) - 1
				b[last] = ':'
			} else {
				ptr := load(ctxptr, code.end.mapPos)
				posPtr := (*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
				*posPtr = append(*posPtr, len(b))
			}
			ptr := load(ctxptr, code.mapIter)
			iter := e.ptrToUnsafePtr(ptr)
			value := mapitervalue(iter)
			store(ctxptr, code.next.idx, uintptr(value))
			mapiternext(iter)
			code = code.next
		case opMapEnd:
			// this operation only used by sorted map.
			length := int(load(ctxptr, code.length))
			type mapKV struct {
				key   string
				value string
			}
			kvs := make([]mapKV, 0, length)
			ptr := load(ctxptr, code.mapPos)
			posPtr := e.ptrToUnsafePtr(ptr)
			pos := *(*[]int)(posPtr)
			for i := 0; i < length; i++ {
				startKey := pos[i*2]
				startValue := pos[i*2+1]
				var endValue int
				if i+1 < length {
					endValue = pos[i*2+2]
				} else {
					endValue = len(b)
				}
				kvs = append(kvs, mapKV{
					key:   string(b[startKey:startValue]),
					value: string(b[startValue:endValue]),
				})
			}
			sort.Slice(kvs, func(i, j int) bool {
				return kvs[i].key < kvs[j].key
			})
			buf := b[pos[0]:]
			buf = buf[:0]
			for _, kv := range kvs {
				buf = append(buf, []byte(kv.key)...)
				buf[len(buf)-1] = ':'
				buf = append(buf, []byte(kv.value)...)
			}
			buf[len(buf)-1] = '}'
			buf = append(buf, ',')
			b = b[:pos[0]]
			b = append(b, buf...)
			code = code.next
		case opMapHeadIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				uptr := e.ptrToUnsafePtr(ptr)
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{', '\n')
					iter := mapiterinit(code.typ, uptr)
					ctx.keepRefs = append(ctx.keepRefs, iter)
					store(ctxptr, code.elemIdx, 0)
					store(ctxptr, code.length, uintptr(mlen))
					store(ctxptr, code.mapIter, uintptr(iter))

					if !e.unorderedMap {
						pos := make([]int, 0, mlen)
						pos = append(pos, len(b))
						posPtr := unsafe.Pointer(&pos)
						ctx.keepRefs = append(ctx.keepRefs, posPtr)
						store(ctxptr, code.end.mapPos, uintptr(posPtr))
					} else {
						b = e.encodeIndent(b, code.next.indent)
					}

					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '{', '}', ',', '\n')
					code = code.end.next
				}
			}
		case opMapHeadLoadIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				code = code.end.next
			} else {
				// load pointer
				ptr = e.ptrToPtr(ptr)
				uptr := e.ptrToUnsafePtr(ptr)
				if uintptr(uptr) == 0 {
					b = e.encodeIndent(b, code.indent)
					b = encodeNull(b)
					b = encodeIndentComma(b)
					code = code.end.next
					break
				}
				mlen := maplen(uptr)
				if mlen > 0 {
					b = append(b, '{', '\n')
					iter := mapiterinit(code.typ, uptr)
					ctx.keepRefs = append(ctx.keepRefs, iter)
					store(ctxptr, code.elemIdx, 0)
					store(ctxptr, code.length, uintptr(mlen))
					store(ctxptr, code.mapIter, uintptr(iter))
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))

					if !e.unorderedMap {
						pos := make([]int, 0, mlen)
						pos = append(pos, len(b))
						posPtr := unsafe.Pointer(&pos)
						ctx.keepRefs = append(ctx.keepRefs, posPtr)
						store(ctxptr, code.end.mapPos, uintptr(posPtr))
					} else {
						b = e.encodeIndent(b, code.next.indent)
					}

					code = code.next
				} else {
					b = e.encodeIndent(b, code.indent)
					b = append(b, '{', '}', ',', '\n')
					code = code.end.next
				}
			}
		case opMapKeyIndent:
			idx := load(ctxptr, code.elemIdx)
			length := load(ctxptr, code.length)
			idx++
			if e.unorderedMap {
				if idx < length {
					b = e.encodeIndent(b, code.indent)
					store(ctxptr, code.elemIdx, idx)
					ptr := load(ctxptr, code.mapIter)
					iter := e.ptrToUnsafePtr(ptr)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					last := len(b) - 1
					b[last] = '\n'
					b = e.encodeIndent(b, code.indent-1)
					b = append(b, '}', ',', '\n')
					code = code.end.next
				}
			} else {
				ptr := load(ctxptr, code.end.mapPos)
				posPtr := (*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
				*posPtr = append(*posPtr, len(b))
				if idx < length {
					ptr := load(ctxptr, code.mapIter)
					iter := e.ptrToUnsafePtr(ptr)
					store(ctxptr, code.elemIdx, idx)
					key := mapiterkey(iter)
					store(ctxptr, code.next.idx, uintptr(key))
					code = code.next
				} else {
					code = code.end
				}
			}
		case opMapValueIndent:
			if e.unorderedMap {
				b = append(b, ':', ' ')
			} else {
				ptr := load(ctxptr, code.end.mapPos)
				posPtr := (*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
				*posPtr = append(*posPtr, len(b))
			}
			ptr := load(ctxptr, code.mapIter)
			iter := e.ptrToUnsafePtr(ptr)
			value := mapitervalue(iter)
			store(ctxptr, code.next.idx, uintptr(value))
			mapiternext(iter)
			code = code.next
		case opMapEndIndent:
			// this operation only used by sorted map
			length := int(load(ctxptr, code.length))
			type mapKV struct {
				key   string
				value string
			}
			kvs := make([]mapKV, 0, length)
			ptr := load(ctxptr, code.mapPos)
			pos := *(*[]int)(*(*unsafe.Pointer)(unsafe.Pointer(&ptr)))
			for i := 0; i < length; i++ {
				startKey := pos[i*2]
				startValue := pos[i*2+1]
				var endValue int
				if i+1 < length {
					endValue = pos[i*2+2]
				} else {
					endValue = len(b)
				}
				kvs = append(kvs, mapKV{
					key:   string(b[startKey:startValue]),
					value: string(b[startValue:endValue]),
				})
			}
			sort.Slice(kvs, func(i, j int) bool {
				return kvs[i].key < kvs[j].key
			})
			buf := b[pos[0]:]
			buf = buf[:0]
			for _, kv := range kvs {
				buf = append(buf, e.prefix...)
				buf = append(buf, bytes.Repeat(e.indentStr, code.indent+1)...)

				buf = append(buf, []byte(kv.key)...)
				buf[len(buf)-2] = ':'
				buf[len(buf)-1] = ' '
				buf = append(buf, []byte(kv.value)...)
			}
			buf = buf[:len(buf)-2]
			buf = append(buf, '\n')
			buf = append(buf, e.prefix...)
			buf = append(buf, bytes.Repeat(e.indentStr, code.indent)...)
			buf = append(buf, '}', ',', '\n')
			b = b[:pos[0]]
			b = append(b, buf...)
			code = code.next
		case opStructFieldPtrAnonymousHeadRecursive:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadRecursive:
			fallthrough
		case opStructFieldRecursive:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				if recursiveLevel > startDetectingCyclesAfter {
					if _, exists := seenPtr[ptr]; exists {
						return nil, errUnsupportedValue(code, ptr)
					}
				}
			}
			seenPtr[ptr] = struct{}{}
			c := code.jmp.code
			c.end.next = newEndOp(&encodeCompileContext{})
			c.op = c.op.ptrHeadToHead()

			beforeLastCode := c.end
			lastCode := beforeLastCode.next

			lastCode.idx = beforeLastCode.idx + uintptrSize
			lastCode.elemIdx = lastCode.idx + uintptrSize

			// extend length to alloc slot for elemIdx
			totalLength := uintptr(code.totalLength() + 1)
			nextTotalLength := uintptr(c.totalLength() + 1)

			curlen := uintptr(len(ctx.ptrs))
			offsetNum := ptrOffset / uintptrSize
			oldOffset := ptrOffset
			ptrOffset += totalLength * uintptrSize

			newLen := offsetNum + totalLength + nextTotalLength
			if curlen < newLen {
				ctx.ptrs = append(ctx.ptrs, make([]uintptr, newLen-curlen)...)
			}
			ctxptr = ctx.ptr() + ptrOffset // assign new ctxptr

			store(ctxptr, c.idx, ptr)
			store(ctxptr, lastCode.idx, oldOffset)
			store(ctxptr, lastCode.elemIdx, uintptr(unsafe.Pointer(code.next)))

			// link lastCode ( opStructFieldRecursiveEnd ) => code.next
			lastCode.op = opStructFieldRecursiveEnd
			code = c
			recursiveLevel++
		case opStructFieldRecursiveEnd:
			recursiveLevel--

			// restore ctxptr
			offset := load(ctxptr, code.idx)
			ptr := load(ctxptr, code.elemIdx)
			code = (*opcode)(e.ptrToUnsafePtr(ptr))
			ctxptr = ctx.ptr() + offset
			ptrOffset = offset
		case opStructFieldPtrHead:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHead {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = e.encodeKey(b, code)
				}
				p := ptr + code.offset
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldAnonymousHead:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructFieldPtrHeadInt:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeInt(b, e.ptrToInt(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeInt(b, e.ptrToInt(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt8:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt8 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeInt8(b, e.ptrToInt8(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeInt8(b, e.ptrToInt8(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt16:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt16 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeInt16(b, e.ptrToInt16(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeInt16(b, e.ptrToInt16(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeInt32(b, e.ptrToInt32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeInt32(b, e.ptrToInt32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadInt64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeInt64(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadInt64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeInt64(b, e.ptrToInt64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeUint(b, e.ptrToUint(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeUint(b, e.ptrToUint(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint8:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint8 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeUint8(b, e.ptrToUint8(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint8:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeUint8(b, e.ptrToUint8(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint16:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint16 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeUint16(b, e.ptrToUint16(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint16:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeUint16(b, e.ptrToUint16(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeUint32(b, e.ptrToUint32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeUint32(b, e.ptrToUint32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadUint64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeUint64(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadUint64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeUint64(b, e.ptrToUint64(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat32:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadFloat32 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadFloat32:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat64:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadFloat64 {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadFloat64:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeKey(b, code)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadString:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadString {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadString:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, e.ptrToString(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadBool:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadBool {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadBool:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeBool(b, e.ptrToBool(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadBytes:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadBytes {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '{', '}', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadBytes:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadArray:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadArray:
			ptr := load(ctxptr, code.idx) + code.offset
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadArray {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '[', ']', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = e.encodeKey(b, code)
				}
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructFieldPtrAnonymousHeadArray:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadArray:
			ptr := load(ctxptr, code.idx) + code.offset
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				store(ctxptr, code.idx, ptr)
				code = code.next
			}
		case opStructFieldPtrHeadSlice:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadSlice:
			ptr := load(ctxptr, code.idx)
			p := ptr + code.offset
			if p == 0 {
				if code.op == opStructFieldPtrHeadSlice {
					b = encodeNull(b)
					b = encodeComma(b)
				} else {
					b = append(b, '[', ']', ',')
				}
				code = code.end.next
			} else {
				b = append(b, '{')
				if !code.anonymousKey {
					b = e.encodeKey(b, code)
				}
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadSlice:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadSlice:
			ptr := load(ctxptr, code.idx)
			p := ptr + code.offset
			if p == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				store(ctxptr, code.idx, p)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalJSON:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					code = code.end
					break
				}
				bb, err := rv.Interface().(Marshaler).MarshalJSON()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				if len(bb) == 0 {
					return nil, errUnexpectedEndOfJSON(
						fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
						0,
					)
				}
				var buf bytes.Buffer
				if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
					return nil, err
				}
				b = append(b, buf.Bytes()...)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadMarshalJSON:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					code = code.end.next
					break
				}
				bb, err := rv.Interface().(Marshaler).MarshalJSON()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				if len(bb) == 0 {
					return nil, errUnexpectedEndOfJSON(
						fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
						0,
					)
				}
				var buf bytes.Buffer
				if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
					return nil, err
				}
				b = append(b, buf.Bytes()...)
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadMarshalText:
			p := load(ctxptr, code.idx)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
				break
			}
			store(ctxptr, code.idx, e.ptrToPtr(p))
			fallthrough
		case opStructFieldHeadMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					b = encodeComma(b)
					code = code.end
					break
				}
				bytes, err := rv.Interface().(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadMarshalText:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldAnonymousHeadMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				ptr += code.offset
				v := e.ptrToInterface(code, ptr)
				rv := reflect.ValueOf(v)
				if rv.Type().Kind() == reflect.Interface && rv.IsNil() {
					b = encodeNull(b)
					b = encodeComma(b)
					code = code.end.next
					break
				}
				bytes, err := rv.Interface().(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else if code.next == code.end {
				// not exists fields
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '}', ',', '\n')
				code = code.end.next
				store(ctxptr, code.idx, ptr)
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, ptr)
			}
		case opStructFieldPtrHeadIntIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				if code.op == opStructFieldPtrHeadIntIndent {
					b = e.encodeIndent(b, code.indent)
					b = encodeNull(b)
					b = encodeIndentComma(b)
				} else {
					b = append(b, '{', '}', ',', '\n')
				}
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt(b, e.ptrToInt(ptr+code.offset))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadInt8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = encodeIndentComma(b)
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt8(b, e.ptrToInt8(ptr))
				b = append(b, ',', '\n')
				code = code.next
			}
		case opStructFieldPtrHeadInt16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt16(b, e.ptrToInt16(ptr))
				b = append(b, ',', '\n')
				code = code.next
			}
		case opStructFieldPtrHeadInt32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt32(b, e.ptrToInt32(ptr))
				b = append(b, ',', '\n')
				code = code.next
			}
		case opStructFieldPtrHeadInt64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt64(b, e.ptrToInt64(ptr))
				b = append(b, ',', '\n')
				code = code.next
			}
		case opStructFieldPtrHeadUintIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint(b, e.ptrToUint(ptr))
				b = append(b, ',', '\n')
				code = code.next
			}
		case opStructFieldPtrHeadUint8Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint8(b, e.ptrToUint8(ptr))
				b = append(b, ',', '\n')
				code = code.next
			}
		case opStructFieldPtrHeadUint16Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint16(b, e.ptrToUint16(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint32(b, e.ptrToUint32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadUint64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint64(b, e.ptrToUint64(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat32Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeFloat32(b, e.ptrToFloat32(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadFloat64Indent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, e.ptrToString(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadBoolIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeBool(b, e.ptrToBool(ptr))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadBytesIndent:
			store(ctxptr, code.idx, e.ptrToPtr(load(ctxptr, code.idx)))
			fallthrough
		case opStructFieldHeadBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				s := base64.StdEncoding.EncodeToString(e.ptrToBytes(ptr))
				b = append(b, '"')
				b = encodeBytes(b, *(*[]byte)(unsafe.Pointer(&s)))
				b = append(b, '"', ',', '\n')
				code = code.next
			}
		case opStructFieldPtrHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				p := ptr + code.offset
				if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmpty:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				p := ptr + code.offset
				if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt8(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt8(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt16(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt16(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt32(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt32(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToInt64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt64(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToInt64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeInt64(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint8(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint8(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint16(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint16(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint32(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint32(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToUint64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint64(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToUint64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeUint64(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToFloat32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeFloat32(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeFloat32(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToFloat64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = e.encodeKey(b, code)
					b = encodeFloat64(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = e.encodeKey(b, code)
					b = encodeFloat64(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToString(ptr + code.offset)
				if v == "" {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = e.encodeString(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToString(ptr + code.offset)
				if v == "" {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = e.encodeString(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToBool(ptr + code.offset)
				if !v {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeBool(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToBool(ptr + code.offset)
				if !v {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeBool(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToBytes(ptr + code.offset)
				if len(v) == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeByteSlice(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToBytes(ptr + code.offset)
				if len(v) == 0 {
					code = code.nextField
				} else {
					b = e.encodeKey(b, code)
					b = encodeByteSlice(b, v)
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bb, err := v.(Marshaler).MarshalJSON()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					if len(bb) == 0 {
						if isPtr {
							return nil, errUnexpectedEndOfJSON(
								fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
								0,
							)
						}
						code = code.nextField
					} else {
						var buf bytes.Buffer
						if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
							return nil, err
						}
						b = e.encodeKey(b, code)
						b = append(b, buf.Bytes()...)
						b = encodeComma(b)
						code = code.next
					}
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bb, err := v.(Marshaler).MarshalJSON()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					if len(bb) == 0 {
						if isPtr {
							return nil, errUnexpectedEndOfJSON(
								fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
								0,
							)
						}
						code = code.nextField
					} else {
						var buf bytes.Buffer
						if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
							return nil, err
						}
						b = e.encodeKey(b, code)
						b = append(b, buf.Bytes()...)
						b = encodeComma(b)
						code = code.next
					}
				}
			}
		case opStructFieldPtrHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bytes, err := v.(encoding.TextMarshaler).MarshalText()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					b = e.encodeKey(b, code)
					b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadOmitEmptyMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				if p == nil || (!isPtr && *(*unsafe.Pointer)(p) == nil) {
					code = code.nextField
				} else {
					v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
					bytes, err := v.(encoding.TextMarshaler).MarshalText()
					if err != nil {
						return nil, &MarshalerError{
							Type: rtype2type(code.typ),
							Err:  err,
						}
					}
					b = e.encodeKey(b, code)
					b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				p := ptr + code.offset
				if p == 0 || *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p))) == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructFieldPtrHeadOmitEmptyIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeInt(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeInt8(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeInt16(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeInt32(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToInt64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeInt64(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeUint(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint8(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeUint8(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint16(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeUint16(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeUint32(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToUint64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeUint64(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToFloat32(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeFloat32(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToFloat64(ptr + code.offset)
				if v == 0 {
					code = code.nextField
				} else {
					if math.IsInf(v, 0) || math.IsNaN(v) {
						return nil, errUnsupportedFloat(v)
					}
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeFloat64(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToString(ptr + code.offset)
				if v == "" {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = e.encodeString(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToBool(ptr + code.offset)
				if !v {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					b = encodeBool(b, v)
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = e.encodeIndent(b, code.indent)
				b = append(b, '{', '\n')
				v := e.ptrToBytes(ptr + code.offset)
				if len(v) == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent+1)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					s := base64.StdEncoding.EncodeToString(v)
					b = append(b, '"')
					b = encodeBytes(b, *(*[]byte)(unsafe.Pointer(&s)))
					b = append(b, '"')
					b = encodeIndentComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				p := ptr + code.offset
				b = e.encodeKey(b, code)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrAnonymousHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTag:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				code = code.next
				store(ctxptr, code.idx, ptr+code.offset)
			}
		case opStructFieldPtrHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt64(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagInt64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt64(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint8:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint8(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint16:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint16(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint64(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagUint64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint64(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToFloat32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat32:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToFloat32(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(v))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagFloat64:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(v))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				var buf bytes.Buffer
				enc := NewEncoder(&buf)
				s := e.ptrToString(ptr + code.offset)
				if e.enabledHTMLEscape {
					enc.buf = encodeEscapedString(enc.buf, s)
				} else {
					enc.buf = encodeNoEscapedString(enc.buf, s)
				}
				b = e.encodeString(b, string(enc.buf))
				b = encodeComma(b)
				enc.release()
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagString:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, strconv.Quote(e.ptrToString(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToBool(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBool:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, fmt.Sprint(e.ptrToBool(ptr+code.offset)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				b = e.encodeKey(b, code)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagBytes:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				b = e.encodeKey(b, code)
				b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(bb) == 0 {
					if isPtr {
						return nil, errUnexpectedEndOfJSON(
							fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
							0,
						)
					}
					b = e.encodeKey(b, code)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
						return nil, err
					}
					b = e.encodeString(b, buf.String())
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalJSON:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				isPtr := code.typ.Kind() == reflect.Ptr
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				if len(bb) == 0 {
					if isPtr {
						return nil, errUnexpectedEndOfJSON(
							fmt.Sprintf("error calling MarshalJSON for type %s", code.typ),
							0,
						)
					}
					b = e.encodeKey(b, code)
					b = append(b, '"', '"')
					b = encodeComma(b)
					code = code.nextField
				} else {
					var buf bytes.Buffer
					if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
						return nil, err
					}
					b = e.encodeKey(b, code)
					b = e.encodeString(b, buf.String())
					b = encodeComma(b)
					code = code.next
				}
			}
		case opStructFieldPtrHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.end.next
			} else {
				b = append(b, '{')
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, &MarshalerError{
						Type: rtype2type(code.typ),
						Err:  err,
					}
				}
				b = e.encodeKey(b, code)
				b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldAnonymousHeadStringTagMarshalText:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				code = code.end.next
			} else {
				ptr += code.offset
				p := e.ptrToUnsafePtr(ptr)
				v := *(*interface{})(unsafe.Pointer(&interfaceHeader{typ: code.typ, ptr: p}))
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = e.encodeKey(b, code)
				b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				p := ptr + code.offset
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldPtrHeadStringTagIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagIntIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagInt64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToInt64(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUintIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint8Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint8(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint16Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint16(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagUint64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToUint64(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat32Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToFloat32(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagFloat64Indent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				v := e.ptrToFloat64(ptr + code.offset)
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(v))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagStringIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				var buf bytes.Buffer
				enc := NewEncoder(&buf)
				s := e.ptrToString(ptr + code.offset)
				if e.enabledHTMLEscape {
					enc.buf = encodeEscapedString(enc.buf, s)
				} else {
					enc.buf = encodeNoEscapedString(enc.buf, s)
				}
				b = e.encodeString(b, string(enc.buf))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBoolIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, fmt.Sprint(e.ptrToBool(ptr+code.offset)))
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructFieldPtrHeadStringTagBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr != 0 {
				store(ctxptr, code.idx, e.ptrToPtr(ptr))
			}
			fallthrough
		case opStructFieldHeadStringTagBytesIndent:
			ptr := load(ctxptr, code.idx)
			if ptr == 0 {
				b = e.encodeIndent(b, code.indent)
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.end.next
			} else {
				b = append(b, '{', '\n')
				b = e.encodeIndent(b, code.indent+1)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				s := base64.StdEncoding.EncodeToString(
					e.ptrToBytes(ptr + code.offset),
				)
				b = append(b, '"')
				b = encodeBytes(b, *(*[]byte)(unsafe.Pointer(&s)))
				b = append(b, '"')
				b = encodeIndentComma(b)
				code = code.next
			}
		case opStructField:
			if !code.anonymousKey {
				b = e.encodeKey(b, code)
			}
			ptr := load(ctxptr, code.headIdx) + code.offset
			code = code.next
			store(ctxptr, code.idx, ptr)
		case opStructFieldPtrInt:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeInt(b, e.ptrToInt(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeInt(b, e.ptrToInt(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrInt8:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeInt8(b, e.ptrToInt8(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeInt8(b, e.ptrToInt8(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrInt16:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeInt16(b, e.ptrToInt16(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeInt16(b, e.ptrToInt16(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrInt32:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeInt32(b, e.ptrToInt32(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeInt32(b, e.ptrToInt32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrInt64:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeInt64(b, e.ptrToInt64(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldInt64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeInt64(b, e.ptrToInt64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrUint:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeUint(b, e.ptrToUint(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeUint(b, e.ptrToUint(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrUint8:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeUint8(b, e.ptrToUint8(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeUint8(b, e.ptrToUint8(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrUint16:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeUint16(b, e.ptrToUint16(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeUint16(b, e.ptrToUint16(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrUint32:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeUint32(b, e.ptrToUint32(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeUint32(b, e.ptrToUint32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrUint64:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeUint64(b, e.ptrToUint64(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldUint64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeUint64(b, e.ptrToUint64(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrFloat32:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeFloat32(b, e.ptrToFloat32(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrFloat64:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
				b = encodeComma(b)
				code = code.next
				break
			}
			v := e.ptrToFloat64(p)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldFloat64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrString:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = e.encodeString(b, e.ptrToString(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldString:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, e.ptrToString(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldPtrBool:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := e.ptrToPtr(ptr + code.offset)
			if p == 0 {
				b = encodeNull(b)
			} else {
				b = encodeBool(b, e.ptrToBool(p))
			}
			b = encodeComma(b)
			code = code.next
		case opStructFieldBool:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldBytes:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = encodeByteSlice(b, e.ptrToBytes(ptr+code.offset))
			b = encodeComma(b)
			code = code.next
		case opStructFieldMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = append(b, buf.Bytes()...)
			b = encodeComma(b)
			code = code.next
		case opStructFieldMarshalText:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldArray:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldSlice:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldMap:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldMapLoad:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldStruct:
			b = e.encodeKey(b, code)
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldIntIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeInt(b, e.ptrToInt(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt8Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeInt8(b, e.ptrToInt8(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt16Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeInt16(b, e.ptrToInt16(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt32Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeInt32(b, e.ptrToInt32(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldInt64Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeInt64(b, e.ptrToInt64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUintIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeUint(b, e.ptrToUint(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint8Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeUint8(b, e.ptrToUint8(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint16Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeUint16(b, e.ptrToUint16(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint32Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeUint32(b, e.ptrToUint32(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldUint64Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeUint64(b, e.ptrToUint64(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat32Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeFloat32(b, e.ptrToFloat32(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldFloat64Indent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = encodeFloat64(b, v)
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeString(b, e.ptrToString(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBoolIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			b = encodeBool(b, e.ptrToBool(ptr+code.offset))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldBytesIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			s := base64.StdEncoding.EncodeToString(e.ptrToBytes(ptr + code.offset))
			b = append(b, '"')
			b = append(b, *(*[]byte)(unsafe.Pointer(&s))...)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldMarshalJSONIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = append(b, buf.Bytes()...)
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldArrayIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructFieldSliceIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				b = encodeNull(b)
				b = encodeIndentComma(b)
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructFieldMapIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				b = encodeNull(b)
				code = code.nextField
			} else {
				p = e.ptrToPtr(p)
				mlen := maplen(e.ptrToUnsafePtr(p))
				if mlen == 0 {
					b = append(b, '{', '}', ',', '\n')
					mapCode := code.next
					code = mapCode.end.next
				} else {
					code = code.next
				}
			}
		case opStructFieldMapLoadIndent:
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				b = encodeNull(b)
				code = code.nextField
			} else {
				p = e.ptrToPtr(p)
				mlen := maplen(e.ptrToUnsafePtr(p))
				if mlen == 0 {
					b = append(b, '{', '}', ',', '\n')
					code = code.nextField
				} else {
					code = code.next
				}
			}
		case opStructFieldStructIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			if p == 0 {
				b = append(b, '{', '}', ',', '\n')
				code = code.nextField
			} else {
				headCode := code.next
				if headCode.next == headCode.end {
					// not exists fields
					b = append(b, '{', '}', ',', '\n')
					code = code.nextField
				} else {
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructFieldOmitEmpty:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || **(**uintptr)(unsafe.Pointer(&p)) == 0 {
				code = code.nextField
			} else {
				b = e.encodeKey(b, code)
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldOmitEmptyInt:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeInt(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeInt8(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeInt16(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeInt32(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeInt64(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeUint(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint8:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeUint8(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint16:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeUint16(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeUint32(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeUint64(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyFloat32:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = e.encodeKey(b, code)
				b = encodeFloat32(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeKey(b, code)
				b = encodeFloat64(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyString:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = e.encodeKey(b, code)
				b = e.encodeString(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyBool:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = e.encodeKey(b, code)
				b = encodeBool(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = e.encodeKey(b, code)
				b = encodeByteSlice(b, v)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			if v != nil {
				bb, err := v.(Marshaler).MarshalJSON()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				var buf bytes.Buffer
				if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
					return nil, err
				}
				b = append(b, buf.Bytes()...)
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			if v != nil {
				bytes, err := v.(encoding.TextMarshaler).MarshalText()
				if err != nil {
					return nil, errMarshaler(code, err)
				}
				b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
				b = encodeComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyArray:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructFieldOmitEmptySlice:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				code = code.nextField
			} else {
				code = code.next
			}
		case opStructFieldOmitEmptyMap:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				mlen := maplen(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
				if mlen == 0 {
					code = code.nextField
				} else {
					code = code.next
				}
			}
		case opStructFieldOmitEmptyMapLoad:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				mlen := maplen(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
				if mlen == 0 {
					code = code.nextField
				} else {
					code = code.next
				}
			}
		case opStructFieldOmitEmptyIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 || **(**uintptr)(unsafe.Pointer(&p)) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				code = code.next
				store(ctxptr, code.idx, p)
			}
		case opStructFieldOmitEmptyIntIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt8(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt16(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt32(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToInt64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeInt64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUintIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint8(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint8(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint16(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint16(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint32(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToUint64(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeUint64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat32(ptr + code.offset)
			if v != 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeFloat32(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyFloat64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if v != 0 {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					return nil, errUnsupportedFloat(v)
				}
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeFloat64(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyStringIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToString(ptr + code.offset)
			if v != "" {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = e.encodeString(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBool(ptr + code.offset)
			if v {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				b = encodeBool(b, v)
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			if len(v) > 0 {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				s := base64.StdEncoding.EncodeToString(v)
				b = append(b, '"')
				b = append(b, *(*[]byte)(unsafe.Pointer(&s))...)
				b = append(b, '"')
				b = encodeIndentComma(b)
			}
			code = code.next
		case opStructFieldOmitEmptyArrayIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			array := e.ptrToSlice(p)
			if p == 0 || uintptr(array.data) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				code = code.next
			}
		case opStructFieldOmitEmptySliceIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			slice := e.ptrToSlice(p)
			if p == 0 || uintptr(slice.data) == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				code = code.next
			}
		case opStructFieldOmitEmptyMapIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				mlen := maplen(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
				if mlen == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					code = code.next
				}
			}
		case opStructFieldOmitEmptyMapLoadIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				mlen := maplen(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
				if mlen == 0 {
					code = code.nextField
				} else {
					b = e.encodeIndent(b, code.indent)
					b = e.encodeKey(b, code)
					b = append(b, ' ')
					code = code.next
				}
			}
		case opStructFieldOmitEmptyStructIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			if p == 0 {
				code = code.nextField
			} else {
				b = e.encodeIndent(b, code.indent)
				b = e.encodeKey(b, code)
				b = append(b, ' ')
				headCode := code.next
				if headCode.next == headCode.end {
					// not exists fields
					b = append(b, '{', '}', ',', '\n')
					code = code.nextField
				} else {
					code = code.next
					store(ctxptr, code.idx, p)
				}
			}
		case opStructFieldStringTag:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeKey(b, code)
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldStringTagInt:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagInt8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagInt16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagInt32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagInt64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt64(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagUint:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagUint8:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint8(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagUint16:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint16(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagUint32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagUint64:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint64(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagFloat32:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToFloat32(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagFloat64:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(v))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagString:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.buf = enc.encodeString(enc.buf, e.ptrToString(ptr+code.offset))
			b = e.encodeString(b, string(enc.buf))
			code = code.next
		case opStructFieldStringTagBool:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeKey(b, code)
			b = e.encodeString(b, fmt.Sprint(e.ptrToBool(ptr+code.offset)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagBytes:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToBytes(ptr + code.offset)
			b = e.encodeKey(b, code)
			b = encodeByteSlice(b, v)
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagMarshalJSON:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = e.encodeString(b, buf.String())
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagMarshalText:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeComma(b)
			code = code.next
		case opStructFieldStringTagIndent:
			ptr := load(ctxptr, code.headIdx)
			p := ptr + code.offset
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			code = code.next
			store(ctxptr, code.idx, p)
		case opStructFieldStringTagIntIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagInt64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToInt64(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUintIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint8Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint8(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint16Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint16(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagUint64Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToUint64(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagFloat32Indent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToFloat32(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagFloat64Indent:
			ptr := load(ctxptr, code.headIdx)
			v := e.ptrToFloat64(ptr + code.offset)
			if math.IsInf(v, 0) || math.IsNaN(v) {
				return nil, errUnsupportedFloat(v)
			}
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(v))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagStringIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.buf = enc.encodeString(enc.buf, e.ptrToString(ptr+code.offset))
			b = e.encodeString(b, string(enc.buf))
			b = encodeIndentComma(b)
			enc.release()
			code = code.next
		case opStructFieldStringTagBoolIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			b = e.encodeString(b, fmt.Sprint(e.ptrToBool(ptr+code.offset)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagBytesIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			s := base64.StdEncoding.EncodeToString(
				e.ptrToBytes(ptr + code.offset),
			)
			b = append(b, '"')
			b = append(b, *(*[]byte)(unsafe.Pointer(&s))...)
			b = append(b, '"')
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagMarshalJSONIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bb, err := v.(Marshaler).MarshalJSON()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			var buf bytes.Buffer
			if err := compact(&buf, bb, e.enabledHTMLEscape); err != nil {
				return nil, err
			}
			b = e.encodeString(b, buf.String())
			b = encodeIndentComma(b)
			code = code.next
		case opStructFieldStringTagMarshalTextIndent:
			ptr := load(ctxptr, code.headIdx)
			b = e.encodeIndent(b, code.indent)
			b = e.encodeKey(b, code)
			b = append(b, ' ')
			p := ptr + code.offset
			v := e.ptrToInterface(code, p)
			bytes, err := v.(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, errMarshaler(code, err)
			}
			b = e.encodeString(b, *(*string)(unsafe.Pointer(&bytes)))
			b = encodeIndentComma(b)
			code = code.next
		case opStructEnd:
			last := len(b) - 1
			if b[last] == ',' {
				b[last] = '}'
			} else {
				b = append(b, '}')
			}
			b = encodeComma(b)
			code = code.next
		case opStructAnonymousEnd:
			code = code.next
		case opStructEndIndent:
			last := len(b) - 1
			if b[last] == '\n' {
				// to remove ',' and '\n' characters
				b = b[:len(b)-2]
			}
			b = append(b, '\n')
			b = e.encodeIndent(b, code.indent)
			b = append(b, '}')
			b = encodeIndentComma(b)
			code = code.next
		case opEnd:
			goto END
		}
	}
END:
	return b, nil
}

func (e *Encoder) ptrToInt(p uintptr) int            { return **(**int)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt8(p uintptr) int8          { return **(**int8)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt16(p uintptr) int16        { return **(**int16)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt32(p uintptr) int32        { return **(**int32)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToInt64(p uintptr) int64        { return **(**int64)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint(p uintptr) uint          { return **(**uint)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint8(p uintptr) uint8        { return **(**uint8)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint16(p uintptr) uint16      { return **(**uint16)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint32(p uintptr) uint32      { return **(**uint32)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToUint64(p uintptr) uint64      { return **(**uint64)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToFloat32(p uintptr) float32    { return **(**float32)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToFloat64(p uintptr) float64    { return **(**float64)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToBool(p uintptr) bool          { return **(**bool)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToByte(p uintptr) byte          { return **(**byte)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToBytes(p uintptr) []byte       { return **(**[]byte)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToString(p uintptr) string      { return **(**string)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToSlice(p uintptr) *sliceHeader { return *(**sliceHeader)(unsafe.Pointer(&p)) }
func (e *Encoder) ptrToPtr(p uintptr) uintptr {
	return uintptr(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
}
func (e *Encoder) ptrToUnsafePtr(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}
func (e *Encoder) ptrToInterface(code *opcode, p uintptr) interface{} {
	return *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: code.typ,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&p)),
	}))
}
