function dec(...args: any[]) {}

abstract class DeclaredAndAbstract {
  @dec
  declare declared: string

  @dec
  abstract abstracted: number

  @dec
  value: boolean
}
