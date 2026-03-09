function dec(...args: any[]) {}

class Ref {}
interface IFace {}
type Alias = string
enum Kind {
  A,
}

class FieldTypes {
  @dec anyType: any
  @dec unknownType: unknown
  @dec objectType: object
  @dec stringType: string
  @dec numberType: number
  @dec booleanType: boolean
  @dec symbolType: symbol
  @dec bigintType: bigint
  @dec literalType: 'x'
  @dec arrayType: string[]
  @dec tupleType: [string, number]
  @dec unionType: string | number
  @dec fnType: (x: string) => number
  @dec classType: Ref
  @dec ifaceType: IFace
  @dec aliasType: Alias
  @dec enumType: Kind
  @dec promiseType: Promise<string>
}
