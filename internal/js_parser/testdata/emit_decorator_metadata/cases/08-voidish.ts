function dec(...args: any[]) {}

class Voidish {
  @dec v: void
  @dec u: undefined
  @dec n: null
  @dec nev: never

  @dec
  method(a: void): void {}
}
