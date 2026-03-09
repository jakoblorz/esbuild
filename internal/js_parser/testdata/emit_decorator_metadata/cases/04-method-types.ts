function dec(...args: any[]) {}

class Ref {}
interface IFace {}
type Alias = string
enum Kind {
  A,
}

class MethodTypes {
  @dec
  rich(
    a: any,
    b: string,
    c: Alias,
    d: IFace,
    e: Kind,
    f: Ref[],
    g: (x: string) => number,
  ): Alias {
    return ''
  }

  @dec
  implicit(value) {}
}
