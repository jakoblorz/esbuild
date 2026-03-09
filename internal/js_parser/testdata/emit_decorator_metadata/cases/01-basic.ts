function dec(...args: any[]) {}

@dec
class Basic {
  constructor(a: string) {}

  @dec
  field: number

  @dec
  method(x: string): boolean {
    return true
  }
}
