function dec(...args: any[]) {}

const key = 'computed'

class ComputedStatic {
  @dec
  static staticValue: Promise<string>

  @dec
  [key]: number
}
