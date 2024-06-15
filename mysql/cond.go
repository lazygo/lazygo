package mysql

import (
	"fmt"
	"strings"
)

type Relate string

func (r Relate) String() string {
	return string(r)
}

const (
	AND Relate = "AND"
	OR  Relate = "OR"
)

type Cond interface {
	String() string
	Args() []any
}

type metaCond struct {
	key  string
	op   string
	args []any
}

func (c *metaCond) String() string {
	if len(c.args) == 0 {
		return fmt.Sprintf("%s IS NULL", buildK(c.key))
	}
	if strings.Contains(strings.ToUpper(c.op), "IN") {
		arr := make([]string, 0, len(c.args))
		for range c.args {
			arr = append(arr, "?")
		}
		return fmt.Sprintf("%s %s(%s)", buildK(c.key), c.op, strings.Join(arr, ", "))
	}
	return build(c.key, c.op)
}

func (c *metaCond) Args() []any {
	return c.args
}

type rawCond struct {
	cond string
	args []any
}

func (c *rawCond) String() string {
	return c.cond
}

func (c *rawCond) Args() []any {
	return c.args
}

type groupCond struct {
	relate Relate
	cond   []Cond
}

func newGroup(relate Relate) *groupCond {
	return &groupCond{
		relate: relate,
		cond:   make([]Cond, 0),
	}
}

func (g *groupCond) sub(relate Relate) *groupCond {
	sub := newGroup(relate)
	g.cond = append(g.cond, sub)
	return sub
}

func (g *groupCond) where(cond ...any) error {

	switch len(cond) {
	case 1:
		switch cond[0].(type) {
		case string, Raw:
			// 字符串查询
			g.whereRaw(cond[0].(string))
		case map[string]any:
			// map拼接查询
			g.whereMap(cond[0].(map[string]any))
		default:
			return ErrInvalidCondArguments
		}
	case 2:
		k, ok := cond[0].(string)
		if !ok {
			return ErrInvalidCondArguments
		}
		// in查询
		in, ok := CreateAnyTypeSlice(cond[1])
		if ok {
			g.meta(k, "IN", in...)
			break
		}
		// k = v
		g.meta(k, "=", cond[1])
	case 3:
		k, ok := cond[0].(string)
		if !ok {
			return ErrInvalidCondArguments
			break
		}
		op, ok := cond[1].(string)
		if !ok {
			return ErrInvalidCondArguments
			break
		}
		if strings.ReplaceAll(strings.ToLower(op), " ", "") == "in" {
			val, ok := CreateAnyTypeSlice(cond[2])
			if !ok {
				val = []any{cond[2]}
			}
			g.meta(k, "IN", val...)
			break
		}
		if strings.ReplaceAll(strings.ToLower(op), " ", "") == "notin" {
			val, ok := CreateAnyTypeSlice(cond[2])
			if !ok {
				val = []any{cond[2]}
			}
			g.meta(k, "NOT IN", val...)
			break
		}
		// k op v
		g.meta(k, op, cond[2])
	default:
		return ErrInvalidCondArguments
	}

	return nil
}

func (g *groupCond) meta(key string, op string, args ...any) {
	g.cond = append(g.cond, &metaCond{
		key:  key,
		op:   op,
		args: args,
	})
}

func (g *groupCond) whereMap(cond map[string]any) {
	for k, v := range cond {
		k = strings.TrimSpace(k)
		if strings.Contains(k, " ") {
			kInfo := strings.SplitN(k, " ", 2)
			g.where(kInfo[0], strings.TrimSpace(kInfo[1]), v)
			continue
		}
		if vv, ok := CreateAnyTypeSlice(v); ok {
			g.meta(k, "IN", vv...)
			continue
		}
		if vv, ok := v.(map[string]any); ok {
			sg := g.sub(Relate(k))
			sg.whereMap(vv)
			continue
		}
		g.meta(k, "=", v)
	}
}

func (g *groupCond) whereRaw(cond string, args ...any) {
	g.cond = append(g.cond, &rawCond{
		cond: strings.TrimSpace(cond),
		args: args,
	})
}

func (g *groupCond) clear() {
	g.cond = make([]Cond, 0)
}

func (g *groupCond) String() string {
	switch len(g.cond) {
	case 0:
		return ""
	case 1:
		return g.cond[0].String()
	}

	var b strings.Builder
	write := func(cond Cond) {
		if sc, ok := cond.(*groupCond); ok && len(sc.cond) > 1 {
			b.WriteString("(")
			b.WriteString(cond.String())
			b.WriteString(")")
			return
		}
		b.WriteString(cond.String())
	}

	write(g.cond[0])
	for _, cond := range g.cond[1:] {
		b.WriteString(" " + g.relate.String() + " ")
		write(cond)
	}
	return b.String()
}

func (g *groupCond) Args() []any {
	var args []any
	for _, cond := range g.cond {
		args = append(args, cond.Args()...)
	}
	return args
}
