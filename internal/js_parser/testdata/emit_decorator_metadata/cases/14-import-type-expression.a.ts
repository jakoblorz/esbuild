declare const test: any

export class A {
  @test
  b: import('./14-import-type-expression.b').B
}
