function dec(...args: any[]) {}

class Ref {}

class InferredField {
  @dec inferred = new Ref()
  @dec text = 'x'
  @dec num = 1
  @dec bool = true
}
