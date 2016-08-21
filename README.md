# go struct initializer

demo project for learning go reflection.

usage:

```go

type MyStruct struct {
	AInt int `default:"10"`
	AStr str `default:"world"`
}

aStruct := MyStruct{}
InitializeStruct(&aStruct)
fmt.Println(aStruct.AInt == 10)
fmt.Println(aStruct.AStr == "world")
```
