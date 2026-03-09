function dec(...args: any[]) {}

class Ref {}

class Generic<T> {
  @dec
  value: T
}

class UnionFiltered {
  @dec a: Ref | null
  @dec b: Ref | undefined
  @dec c: null | undefined
  @dec d: 'foo' | 'bar'
  @dec e: true | false
}
