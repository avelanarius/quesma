package transform

import (
	"fmt"
	"mitmproxy/quesma/eql/parser"
	"strconv"
	"strings"
)

type EQLParseTreeToExpTransformer struct {
	parser.BaseEQLVisitor

	// category field name can be customized
	// it's provided as a parameter in the query
	// default is "category.name"
	CategoryFieldName string

	Errors []string
}

func NewEQLParseTreeToExpTransformer() *EQLParseTreeToExpTransformer {
	return &EQLParseTreeToExpTransformer{
		CategoryFieldName: "category.name",
	}
}

func (v *EQLParseTreeToExpTransformer) error(msg string) {
	v.Errors = append(v.Errors, msg)
}

func (v *EQLParseTreeToExpTransformer) evalString(s string) string {

	const tripleQuote = `"""`
	if strings.HasPrefix(s, tripleQuote) && strings.HasSuffix(s, tripleQuote) {
		return s[3 : len(s)-3]
	}

	const quote = `"`
	if strings.HasPrefix(s, quote) && strings.HasSuffix(s, quote) {
		// TODO handle escape sequences
		return s[1 : len(s)-1]
	}

	return s
}

func (v *EQLParseTreeToExpTransformer) evalNumber(s string) (int, error) {
	return strconv.Atoi(s)
}

func (v *EQLParseTreeToExpTransformer) VisitQuery(ctx *parser.QueryContext) interface{} {
	return ctx.SimpleQuery().Accept(v)
}

func (v *EQLParseTreeToExpTransformer) VisitSimpleQuery(ctx *parser.SimpleQueryContext) interface{} {

	category := ctx.Category().Accept(v)
	condition := ctx.Condition().Accept(v)

	if condition == nil {
		if category == nil {
			return nil // TODO what is an empty query?  -> select * from where true
		} else {
			return category
		}
	} else {
		if category != nil {
			return NewInfixOp("and", condition.(Exp), category.(Exp))
		} else {
			return condition
		}
	}
}

func (v *EQLParseTreeToExpTransformer) VisitConditionBoolean(ctx *parser.ConditionBooleanContext) interface{} {
	return v.evalBoolean(ctx.GetText())
}

func (v *EQLParseTreeToExpTransformer) VisitConditionLogicalOp(ctx *parser.ConditionLogicalOpContext) interface{} {
	left := ctx.GetLeft().Accept(v)
	right := ctx.GetRight().Accept(v)
	op := strings.ToLower(ctx.GetOp().GetText())

	return NewInfixOp(op, left.(Exp), right.(Exp))
}

func (v *EQLParseTreeToExpTransformer) VisitConditionOp(ctx *parser.ConditionOpContext) interface{} {

	field := ctx.Field().Accept(v)
	value := ctx.Value().Accept(v)
	op := ctx.GetOp().GetText()

	// paranoia check, should never happen
	// if there is no visitor implemented for the right side value is null

	// TODO add more info here to help debugging
	if value == nil {
		v.error("value is nil here")
		return &Const{Value: "error"}
	}

	return NewInfixOp(op, field.(Exp), value.(Exp))
}

func (v *EQLParseTreeToExpTransformer) VisitConditionOpList(ctx *parser.ConditionOpListContext) interface{} {
	field := ctx.Field().Accept(v)
	op := ctx.GetOp().GetText()
	op = strings.ToLower(op)
	inList := ctx.GetList().Accept(v).(Exp)

	return NewInfixOp(op, field.(Exp), inList)
}

func (v *EQLParseTreeToExpTransformer) VisitConditionNot(ctx *parser.ConditionNotContext) interface{} {
	inner := ctx.Condition().Accept(v)
	return NewPrefixOp("not", []Exp{inner.(Exp)})
}

func (v *EQLParseTreeToExpTransformer) VisitConditionGroup(ctx *parser.ConditionGroupContext) interface{} {
	return NewGroup(ctx.Condition().Accept(v).(Exp))
}

func (v *EQLParseTreeToExpTransformer) VisitConditionNotIn(ctx *parser.ConditionNotInContext) interface{} {
	field := ctx.Field().Accept(v).(Exp)
	inList := ctx.GetList().Accept(v).(Exp)

	return NewInfixOp("not in", field, inList)
}

func (v *EQLParseTreeToExpTransformer) VisitConditionFuncall(ctx *parser.ConditionFuncallContext) interface{} {

	return ctx.Funcall().Accept(v)
}

func (v *EQLParseTreeToExpTransformer) VisitField(ctx *parser.FieldContext) interface{} {

	name := v.evalString(ctx.GetText())

	return NewSymbol(name)

}

func (v *EQLParseTreeToExpTransformer) VisitValueLiteral(ctx *parser.ValueLiteralContext) interface{} {

	return ctx.Literal().Accept(v)
}

func (v *EQLParseTreeToExpTransformer) VisitLiteral(ctx *parser.LiteralContext) interface{} {
	switch {

	case ctx.STRING() != nil:
		return &Const{Value: v.evalString(ctx.GetText())}
	case ctx.NUMBER() != nil:
		i, err := v.evalNumber(ctx.GetText())
		if err == nil {
			return &Const{Value: i}
		}

		v.error(fmt.Sprintf("error parsing number: %v", err))
		return &Const{Value: 0}

	case ctx.BOOLEAN() != nil:
		return v.evalBoolean(ctx.GetText())
	}

	return nil
}

func (v *EQLParseTreeToExpTransformer) VisitValueGroup(ctx *parser.ValueGroupContext) interface{} {

	return NewGroup(ctx.Value().Accept(v).(Exp))

}

func (v *EQLParseTreeToExpTransformer) VisitValueFuncall(ctx *parser.ValueFuncallContext) interface{} {
	return ctx.Funcall().Accept(v)
}

func (v *EQLParseTreeToExpTransformer) VisitValueAddSub(ctx *parser.ValueAddSubContext) interface{} {
	left := ctx.GetLeft().Accept(v)
	right := ctx.GetRight().Accept(v)
	op := ctx.GetOp().GetText()
	return NewInfixOp(op, left.(Exp), right.(Exp))
}

func (v *EQLParseTreeToExpTransformer) VisitValueMulDiv(ctx *parser.ValueMulDivContext) interface{} {
	left := ctx.GetLeft().Accept(v)
	right := ctx.GetRight().Accept(v)
	op := ctx.GetOp().GetText()
	return NewInfixOp(op, left.(Exp), right.(Exp))
}

func (v *EQLParseTreeToExpTransformer) VisitValueField(ctx *parser.ValueFieldContext) interface{} {
	return ctx.Field().Accept(v)
}

func (v *EQLParseTreeToExpTransformer) VisitValueNull(ctx *parser.ValueNullContext) interface{} {
	return NULL
}

func (v *EQLParseTreeToExpTransformer) VisitFuncall(ctx *parser.FuncallContext) interface{} {

	name := ctx.FuncName().Accept(v).(string)

	var args []Exp

	for _, a := range ctx.AllValue() {
		args = append(args, a.Accept(v).(Exp))
	}

	return NewFunction(name, args)

}

func (v *EQLParseTreeToExpTransformer) VisitFuncName(ctx *parser.FuncNameContext) interface{} {
	return ctx.GetText()
}

func (v *EQLParseTreeToExpTransformer) VisitLiteralList(ctx *parser.LiteralListContext) interface{} {

	var values []Exp

	for _, l := range ctx.AllLiteral() {
		values = append(values, l.Accept(v).(Exp))
	}

	return NewArray(values...)
}

func (v *EQLParseTreeToExpTransformer) VisitCategory(ctx *parser.CategoryContext) interface{} {

	var category string
	switch {
	case ctx.ID() != nil:
		category = ctx.ID().GetText()
	case ctx.STRING() != nil:
		category = v.evalString(ctx.STRING().GetText())
	case ctx.ANY() != nil:
	default:
	}

	if category != "" {
		return NewInfixOp("==", NewSymbol(v.CategoryFieldName), NewConst(category))
	}
	// match all
	return nil
}

func (v *EQLParseTreeToExpTransformer) evalBoolean(s string) Exp {
	if strings.ToLower(s) == "true" {
		return TRUE
	}

	return FALSE
}