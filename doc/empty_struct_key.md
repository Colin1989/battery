# Interesting Empty Struct Context Key

Talk is cheap, show me the code:
```go
  type myKey struct{}
  ctx = context.WithValue(ctx, myKey{}, "my value") // Set value
  myValue, ok := ctx.Value(myKey{}}).(string) // Get value

  //ok = true
```

We realise that the context is using an empty struct as a key to the context values. 
So why we are not using a constant string or a large integer as the key instead?

This turned out to be a very insightful engineering implementation. 
We need to know whe an empty struct can be an unique key and why it is better than 
using other keys in context.

## How does interface comparator works here

Firstly we write an interface comparator. 
This is because the `context.WithValue()` takes in interfaces as key and values.

```go
// This function compares the underlying interface type and value pointers. 
//Note that the underlying values are compared depending on the dynamic type of the interface.
func CompareInterfaces(key1 interface{}, key2 interface{}) {
  type ifaceHdr struct {
    T unsafe.Pointer  
    V unsafe.Pointer
  }

  fmt.Println("\ninterface1 == interface2?", key1 == key2)  
  fmt.Printf("interface1 %+v\n", *(*ifaceHdr)(unsafe.Pointer(&key1)))
  fmt.Printf("interface2 %+v\n", *(*ifaceHdr)(unsafe.Pointer(&key2)))
}
```

Then we define two empty struct types `StructKey1` and `StructKey2` as well as two int types

```go
type StructKey1 struct{}
type StructKey2 struct{}

type IntKey1 int
type IntKey2 int

func CompareStruct()  {
  a := "dontUseString"
  b := "dontUseString"
  utils.CompareInterfaces(a, b) // true
  utils.CompareInterfaces(StructKey1{}, StructKey1{}) // true
  utils.CompareInterfaces(StructKey1{}, StructKey2{}) // false
  utils.CompareInterfaces(IntKey1(0), IntKey1(0)) // true
  utils.CompareInterfaces(IntKey1(0), IntKey2(0)) // false
  utils.CompareInterfaces(IntKey1(0), IntKey1(1)) // false
  utils.CompareInterfaces(IntKey1(0), 0) // false
}
```

Let see the comparison result

```go
//a := "dontUseString"
//b := "dontUseString"
//CompareInterfaces(a, b)
interface1 == interface2? true
interface1 {T:0x10b0b00 V:0xc000096240}
interface2 {T:0x10b0b00 V:0xc000096250}
```

Though they have different value pointers the underlying string has same characters so its true. 
However string comparison takes time to traverse the string.

```go
//CompareInterfaces(StructKey1{}, StructKey1{})
interface1 == interface2? true
interface1 {T:0x10b4c60 V:0x11913f0}
interface2 {T:0x10b4c60 V:0x11913f0}
```

Creating two new `StructKey1{}` s and they are empty structs. 
However we see that they have the same type as well as same value pointer? 
Why is it like that since they are different struct instances?

```go
//utils.CompareInterfaces(StructKey1{}, StructKey2{})
interface1 == interface2? false
interface1 {T:0x10b4c60 V:0x11913f0}
interface2 {T:0x10b4cc0 V:0x11913f0}
```

`StructKey1{}` and `StructKey2{}` are empty structs they have the same type as well as same value pointer? 
This is because:

>> Empty struct type instances do not invoke memory allocation and they all point to the same place, 
>> where it occupies 0 byte and it is called runtime.zerobase. 
>> This is because when memory allocation for 0 byte structs this code will be called:

```go
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
  ...
  if size == 0 {
    return unsafe.Pointer(&zerobase)
  }
  ...
}
```

Now we understand what is going on in the new empty struct memory allocation process. 
They all equal to the `zerobase` pointer with different type predefined in the code and thats 
why two new `StructKey1{}` instances are equal and `StructKey1{}` does not equal to `StructKey2{}`

```go
//utils.CompareInterfaces(StructKey1{}, StructKey2{})
utils.CompareInterfaces(IntKey1(0), IntKey1(0))
interface1 == interface2? true
interface1 {T:0x10afe00 V:0x10e9ab0}
interface2 {T:0x10afe00 V:0x10e9ab0}
```

Comparing predefined integer type with same values are true.

```go
utils.CompareInterfaces(IntKey1(0), IntKey1(0)) // true
utils.CompareInterfaces(IntKey1(0), IntKey2(0)) // false
utils.CompareInterfaces(IntKey1(0), IntKey1(1)) // false
```

Comparing predefined integer type with same values are true. 
Comparing different type integer with same int value will give you false. 
Comparing same integer type with different int value will give you false also.

```go
utils.CompareInterfaces(IntKey1(0), 0)
interface1 == interface2? false
interface1 {T:0x10afe00 V:0x10e9ab0}
interface2 {T:0x10b0440 V:0x10e9ab0}
```

And it is interesting that integer literals are not of same type as defined in the code. 
It looks confusing but the underlying logic is that, 
the type has to be the same and dynamic value has to be the same. 
Struct instances also compare their underlying fields 
if they are comparable however in this case empty struct occupy zero space and 
directly assigned the zerobase memory point and therefore it is same for all instances.****

## Why empty struct instance is a good key

Now we come back to the question why they put empty struct instance as the context key. 
The background is that the context may pass around by RPC calls among our services and 
everyone has the ability to edit the context. 
So it is not safe to use string literals or integer literals as keys as they do not have unique type. 
A uuid can be a key if the uuid is a predefined int64 type called type `WhateverUUID` int64 and 
the type info would segregate the value info in the context and this avoids duplication.

Secondly, since in most cases for an interface key, 
we can neglect the value field of the interface and just take out the type information and it is unique enough for key. 
In this case an empty struct is better than other types because they dont cause memory allocation and 
it is faster and more efficient than any other types! 
(If you are tailing about a lot of staff to put into the context with a lot of UUIDs as key. 
You should put these in a map and put the map as the value of a context key. 
You should not put them in the context on by one).

So do use empty structs as context keys when passing it around through RPC calls!