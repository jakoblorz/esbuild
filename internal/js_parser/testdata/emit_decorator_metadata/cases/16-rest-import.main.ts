import { SomeClass } from './16-rest-import.aux'

declare const annotation: any
declare const annotation1: any

@annotation
class ClassA {
  constructor(...init: SomeClass[]) {}

  @annotation1
  foo(...args: SomeClass[]) {}
}
