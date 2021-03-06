package spectests

import (
	"fmt"

	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"

	"math"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"

	"github.com/influxdata/flux/semantic"
)

func init() {
	RegisterFixture(
		AggregateTest(func(aggregate flux.Operation) (stmt string, spec *flux.Spec) {
			return fmt.Sprintf(`SELECT %s(value) FROM db0..cpu WHERE time >= now() - 10m GROUP BY time(5m, 12m)`, aggregate.Spec.Kind()),
				&flux.Spec{
					Operations: []*flux.Operation{
						{
							ID: "from0",
							Spec: &inputs.FromOpSpec{
								BucketID: bucketID.String(),
							},
						},
						{
							ID: "range0",
							Spec: &transformations.RangeOpSpec{
								Start:       flux.Time{Absolute: Now().Add(-10 * time.Minute)},
								Stop:        flux.Time{Absolute: Now()},
								TimeColumn:  execute.DefaultTimeColLabel,
								StartColumn: execute.DefaultStartColLabel,
								StopColumn:  execute.DefaultStopColLabel,
							},
						},
						{
							ID: "filter0",
							Spec: &transformations.FilterOpSpec{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{
												{Key: &semantic.Identifier{Name: "r"}},
											},
										},
										Body: &semantic.LogicalExpression{
											Operator: ast.AndOperator,
											Left: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object: &semantic.IdentifierExpression{
														Name: "r",
													},
													Property: "_measurement",
												},
												Right: &semantic.StringLiteral{
													Value: "cpu",
												},
											},
											Right: &semantic.BinaryExpression{
												Operator: ast.EqualOperator,
												Left: &semantic.MemberExpression{
													Object: &semantic.IdentifierExpression{
														Name: "r",
													},
													Property: "_field",
												},
												Right: &semantic.StringLiteral{
													Value: "value",
												},
											},
										},
									},
								},
							},
						},
						{
							ID: "group0",
							Spec: &transformations.GroupOpSpec{
								Columns: []string{"_measurement", "_start"},
								Mode:    "by",
							},
						},
						{
							ID: "window0",
							Spec: &transformations.WindowOpSpec{
								Every:       flux.Duration(5 * time.Minute),
								Period:      flux.Duration(5 * time.Minute),
								Start:       flux.Time{Absolute: time.Unix(0, 0).Add(time.Minute * 2)},
								TimeColumn:  execute.DefaultTimeColLabel,
								StartColumn: execute.DefaultStartColLabel,
								StopColumn:  execute.DefaultStopColLabel,
							},
						},
						&aggregate,
						{
							ID: "duplicate0",
							Spec: &transformations.DuplicateOpSpec{
								Column: execute.DefaultStartColLabel,
								As:     execute.DefaultTimeColLabel,
							},
						},
						{
							ID: "window1",
							Spec: &transformations.WindowOpSpec{
								Every:       flux.Duration(math.MaxInt64),
								Period:      flux.Duration(math.MaxInt64),
								TimeColumn:  execute.DefaultTimeColLabel,
								StartColumn: execute.DefaultStartColLabel,
								StopColumn:  execute.DefaultStopColLabel,
							},
						},
						{
							ID: "map0",
							Spec: &transformations.MapOpSpec{
								Fn: &semantic.FunctionExpression{
									Block: &semantic.FunctionBlock{
										Parameters: &semantic.FunctionParameters{
											List: []*semantic.FunctionParameter{{
												Key: &semantic.Identifier{Name: "r"},
											}},
										},
										Body: &semantic.ObjectExpression{
											Properties: []*semantic.Property{
												{
													Key: &semantic.Identifier{Name: "_time"},
													Value: &semantic.MemberExpression{
														Object: &semantic.IdentifierExpression{
															Name: "r",
														},
														Property: "_time",
													},
												},
												{
													Key: &semantic.Identifier{Name: string(aggregate.Spec.Kind())},
													Value: &semantic.MemberExpression{
														Object: &semantic.IdentifierExpression{
															Name: "r",
														},
														Property: "_value",
													},
												},
											},
										},
									},
								},
								MergeKey: true,
							},
						},
						{
							ID: "yield0",
							Spec: &transformations.YieldOpSpec{
								Name: "0",
							},
						},
					},
					Edges: []flux.Edge{
						{Parent: "from0", Child: "range0"},
						{Parent: "range0", Child: "filter0"},
						{Parent: "filter0", Child: "group0"},
						{Parent: "group0", Child: "window0"},
						{Parent: "window0", Child: aggregate.ID},
						{Parent: aggregate.ID, Child: "duplicate0"},
						{Parent: "duplicate0", Child: "window1"},
						{Parent: "window1", Child: "map0"},
						{Parent: "map0", Child: "yield0"},
					},
					Now: Now(),
				}
		}),
	)
}
