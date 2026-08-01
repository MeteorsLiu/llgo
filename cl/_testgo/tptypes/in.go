// LITTEST
package main

// CHECK: {{^}}@0 = private unnamed_addr constant [5 x i8] c"hello", align 1{{$}}

// CHECK-LABEL: define void @main.init(){{.*}} {
// CHECK-NEXT: _llgo_0:
// CHECK-NEXT:   %0 = load i1, ptr @"main.init$guard", align 1
// CHECK-NEXT:   br i1 %0, label %_llgo_2, label %_llgo_1
// CHECK-EMPTY:
// CHECK-NEXT: _llgo_1:                                          ; preds = %_llgo_0
// CHECK-NEXT:   store i1 true, ptr @"main.init$guard", align 1
// CHECK-NEXT:   br label %_llgo_2
// CHECK-EMPTY:
// CHECK-NEXT: _llgo_2:                                          ; preds = %_llgo_1, %_llgo_0
// CHECK-NEXT:   ret void
// CHECK-NEXT: }

type Data[T any] struct {
	v T
}

func (p *Data[T]) Set(v T) {
	p.v = v
}

func (p *(Data[T1])) Set2(v T1) {
	p.v = v
}

type sliceOf[E any] interface {
	~[]E
}

type Slice[S sliceOf[T], T any] struct {
	Data S
}

func (p *Slice[S, T]) Append(t ...T) S {
	p.Data = append(p.Data, t...)
	return p.Data
}

func (p *Slice[S1, T1]) Append2(t ...T1) S1 {
	p.Data = append(p.Data, t...)
	return p.Data
}

type (
	DataInt     = Data[int]
	SliceInt    = Slice[[]int, int]
	DataString  = Data[string]
	SliceString = Slice[[]string, string]
)

// CHECK-LABEL: define void @main.main(){{.*}} {
// CHECK-NEXT: _llgo_0:
// CHECK-NEXT:   %0 = alloca %"main.Data[int]", align 8
// CHECK-NEXT:   call void @llvm.memset.p0.i64(ptr %0, i8 0, i64 8, i1 false)
// CHECK-NEXT:   %1 = getelementptr inbounds %"main.Data[int]", ptr %0, i32 0, i32 0
// CHECK-NEXT:   store i64 1, ptr %1, align 8
// CHECK-NEXT:   %2 = load %"main.Data[int]", ptr %0, align 8
// CHECK-NEXT:   %3 = extractvalue %"main.Data[int]" %2, 0
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintInt"(i64 %3)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   %4 = alloca %"main.Data[string]", align 8
// CHECK-NEXT:   call void @llvm.memset.p0.i64(ptr %4, i8 0, i64 16, i1 false)
// CHECK-NEXT:   %5 = getelementptr inbounds %"main.Data[string]", ptr %4, i32 0, i32 0
// CHECK-NEXT:   store %"{{.*}}/runtime/internal/runtime.String" { ptr @0, i64 5 }, ptr %5, align 8
// CHECK-NEXT:   %6 = load %"main.Data[string]", ptr %4, align 8
// CHECK-NEXT:   %7 = extractvalue %"main.Data[string]" %6, 0
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintString"(%"{{.*}}/runtime/internal/runtime.String" %7)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   %8 = alloca %"main.Data[int]", align 8
// CHECK-NEXT:   call void @llvm.memset.p0.i64(ptr %8, i8 0, i64 8, i1 false)
// CHECK-NEXT:   %9 = getelementptr inbounds %"main.Data[int]", ptr %8, i32 0, i32 0
// CHECK-NEXT:   store i64 100, ptr %9, align 8
// CHECK-NEXT:   %10 = load %"main.Data[int]", ptr %8, align 8
// CHECK-NEXT:   %11 = extractvalue %"main.Data[int]" %10, 0
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintInt"(i64 %11)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   %12 = alloca %"main.Data[string]", align 8
// CHECK-NEXT:   call void @llvm.memset.p0.i64(ptr %12, i8 0, i64 16, i1 false)
// CHECK-NEXT:   %13 = getelementptr inbounds %"main.Data[string]", ptr %12, i32 0, i32 0
// CHECK-NEXT:   store %"{{.*}}/runtime/internal/runtime.String" { ptr @0, i64 5 }, ptr %13, align 8
// CHECK-NEXT:   %14 = load %"main.Data[string]", ptr %12, align 8
// CHECK-NEXT:   %15 = extractvalue %"main.Data[string]" %14, 0
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintString"(%"{{.*}}/runtime/internal/runtime.String" %15)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintInt"(i64 0)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   %.stack = alloca i8, i64 24, align 8
// CHECK-NEXT:   call void @llvm.memset.p0.i64(ptr %.stack, i8 0, i64 24, i1 false)
// CHECK-NEXT:   %16 = call ptr @"{{.*}}/runtime/internal/runtime.AllocZ"(i64 8)
// CHECK-NEXT:   %17 = getelementptr inbounds i64, ptr %16, i64 0
// CHECK-NEXT:   store i64 100, ptr %17, align 8
// CHECK-NEXT:   %18 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" undef, ptr %16, 0
// CHECK-NEXT:   %19 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %18, i64 1, 1
// CHECK-NEXT:   %20 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %19, i64 1, 2
// CHECK-NEXT:   %21 = call %"{{.*}}/runtime/internal/runtime.Slice" @"main.(*Slice{{\[\[\]int,int\]}}).Append"(ptr %.stack, %"{{.*}}/runtime/internal/runtime.Slice" %20)
// CHECK-NEXT:   %.stack1 = alloca i8, i64 24, align 8
// CHECK-NEXT:   call void @llvm.memset.p0.i64(ptr %.stack1, i8 0, i64 24, i1 false)
// CHECK-NEXT:   %22 = call ptr @"{{.*}}/runtime/internal/runtime.AllocZ"(i64 16)
// CHECK-NEXT:   %23 = getelementptr inbounds %"{{.*}}/runtime/internal/runtime.String", ptr %22, i64 0
// CHECK-NEXT:   store %"{{.*}}/runtime/internal/runtime.String" { ptr @0, i64 5 }, ptr %23, align 8
// CHECK-NEXT:   %24 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" undef, ptr %22, 0
// CHECK-NEXT:   %25 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %24, i64 1, 1
// CHECK-NEXT:   %26 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %25, i64 1, 2
// CHECK-NEXT:   %27 = call %"{{.*}}/runtime/internal/runtime.Slice" @"main.(*Slice{{\[\[\]string,string\]}}).Append"(ptr %.stack1, %"{{.*}}/runtime/internal/runtime.Slice" %26)
// CHECK-NEXT:   %.stack2 = alloca i8, i64 24, align 8
// CHECK-NEXT:   call void @llvm.memset.p0.i64(ptr %.stack2, i8 0, i64 24, i1 false)
// CHECK-NEXT:   %28 = call ptr @"{{.*}}/runtime/internal/runtime.AllocZ"(i64 32)
// CHECK-NEXT:   %29 = getelementptr inbounds i64, ptr %28, i64 0
// CHECK-NEXT:   store i64 1, ptr %29, align 8
// CHECK-NEXT:   %30 = getelementptr inbounds i64, ptr %28, i64 1
// CHECK-NEXT:   store i64 2, ptr %30, align 8
// CHECK-NEXT:   %31 = getelementptr inbounds i64, ptr %28, i64 2
// CHECK-NEXT:   store i64 3, ptr %31, align 8
// CHECK-NEXT:   %32 = getelementptr inbounds i64, ptr %28, i64 3
// CHECK-NEXT:   store i64 4, ptr %32, align 8
// CHECK-NEXT:   %33 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" undef, ptr %28, 0
// CHECK-NEXT:   %34 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %33, i64 4, 1
// CHECK-NEXT:   %35 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %34, i64 4, 2
// CHECK-NEXT:   %36 = call %"{{.*}}/runtime/internal/runtime.Slice" @"main.(*Slice{{\[\[\]int,int\]}}).Append"(ptr %.stack2, %"{{.*}}/runtime/internal/runtime.Slice" %35)
// CHECK-NEXT:   %37 = call ptr @"{{.*}}/runtime/internal/runtime.AllocZ"(i64 32)
// CHECK-NEXT:   %38 = getelementptr inbounds i64, ptr %37, i64 0
// CHECK-NEXT:   store i64 1, ptr %38, align 8
// CHECK-NEXT:   %39 = getelementptr inbounds i64, ptr %37, i64 1
// CHECK-NEXT:   store i64 2, ptr %39, align 8
// CHECK-NEXT:   %40 = getelementptr inbounds i64, ptr %37, i64 2
// CHECK-NEXT:   store i64 3, ptr %40, align 8
// CHECK-NEXT:   %41 = getelementptr inbounds i64, ptr %37, i64 3
// CHECK-NEXT:   store i64 4, ptr %41, align 8
// CHECK-NEXT:   %42 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" undef, ptr %37, 0
// CHECK-NEXT:   %43 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %42, i64 4, 1
// CHECK-NEXT:   %44 = insertvalue %"{{.*}}/runtime/internal/runtime.Slice" %43, i64 4, 2
// CHECK-NEXT:   %45 = call %"{{.*}}/runtime/internal/runtime.Slice" @"main.(*Slice{{\[\[\]int,int\]}}).Append2"(ptr %.stack2, %"{{.*}}/runtime/internal/runtime.Slice" %44)
// CHECK-NEXT:   %46 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %.stack, i32 0, i32 0
// CHECK-NEXT:   %47 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %46, align 8
// CHECK-NEXT:   %48 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %.stack, i32 0, i32 0
// CHECK-NEXT:   %49 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %48, align 8
// CHECK-NEXT:   %50 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %49, 0
// CHECK-NEXT:   %51 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %49, 1
// CHECK-NEXT:   %52 = icmp uge i64 0, %51
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.CheckIndexRange"(i1 %52, {{.*}})
// CHECK-NEXT:   %53 = getelementptr inbounds i64, ptr %50, i64 0
// CHECK-NEXT:   %54 = load i64, ptr %53, align 8
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintSlice"(%"{{.*}}/runtime/internal/runtime.Slice" %47)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 32)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintInt"(i64 %54)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   %55 = getelementptr inbounds %"main.Slice{{\[\[\]string,string\]}}", ptr %.stack1, i32 0, i32 0
// CHECK-NEXT:   %56 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %55, align 8
// CHECK-NEXT:   %57 = getelementptr inbounds %"main.Slice{{\[\[\]string,string\]}}", ptr %.stack1, i32 0, i32 0
// CHECK-NEXT:   %58 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %57, align 8
// CHECK-NEXT:   %59 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %58, 0
// CHECK-NEXT:   %60 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %58, 1
// CHECK-NEXT:   %61 = icmp uge i64 0, %60
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.CheckIndexRange"(i1 %61, {{.*}})
// CHECK-NEXT:   %62 = getelementptr inbounds %"{{.*}}/runtime/internal/runtime.String", ptr %59, i64 0
// CHECK-NEXT:   %63 = load %"{{.*}}/runtime/internal/runtime.String", ptr %62, align 8
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintSlice"(%"{{.*}}/runtime/internal/runtime.Slice" %56)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 32)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintString"(%"{{.*}}/runtime/internal/runtime.String" %63)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   %64 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %.stack2, i32 0, i32 0
// CHECK-NEXT:   %65 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %64, align 8
// CHECK-NEXT:   %66 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %.stack2, i32 0, i32 0
// CHECK-NEXT:   %67 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %66, align 8
// CHECK-NEXT:   %68 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %67, 0
// CHECK-NEXT:   %69 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %67, 1
// CHECK-NEXT:   %70 = icmp uge i64 0, %69
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.CheckIndexRange"(i1 %70, {{.*}})
// CHECK-NEXT:   %71 = getelementptr inbounds i64, ptr %68, i64 0
// CHECK-NEXT:   %72 = load i64, ptr %71, align 8
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintSlice"(%"{{.*}}/runtime/internal/runtime.Slice" %65)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 32)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintInt"(i64 %72)
// CHECK-NEXT:   call void @"{{.*}}/runtime/internal/runtime.PrintByte"(i8 10)
// CHECK-NEXT:   ret void
// CHECK-NEXT: }

func main() {
	println(DataInt{1}.v)
	println(DataString{"hello"}.v)
	println(Data[int]{100}.v)
	println(Data[string]{"hello"}.v)

	// TODO
	println(Data[struct {
		X int
		Y int
	}]{}.v.X)

	v1 := SliceInt{}
	v1.Append(100)
	v2 := SliceString{}
	v2.Append("hello")
	v3 := Slice[[]int, int]{}
	v3.Append([]int{1, 2, 3, 4}...)
	v3.Append2([]int{1, 2, 3, 4}...)

	println(v1.Data, v1.Data[0])
	println(v2.Data, v2.Data[0])
	println(v3.Data, v3.Data[0])
}

// CHECK-LABEL: define linkonce %"{{.*}}/runtime/internal/runtime.Slice" @"main.(*Slice{{\[\[\]int,int\]}}).Append"(ptr %0, %"{{.*}}/runtime/internal/runtime.Slice" %1){{.*}} {
// CHECK-NEXT: _llgo_0:
// CHECK-NEXT:   %2 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   %3 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %2, align 8
// CHECK-NEXT:   %4 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %1, 0
// CHECK-NEXT:   %5 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %1, 1
// CHECK-NEXT:   %6 = call %"{{.*}}/runtime/internal/runtime.Slice" @"{{.*}}/runtime/internal/runtime.SliceAppend"(%"{{.*}}/runtime/internal/runtime.Slice" %3, ptr %4, i64 %5, i64 8)
// CHECK-NEXT:   %7 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   store %"{{.*}}/runtime/internal/runtime.Slice" %6, ptr %7, align 8
// CHECK-NEXT:   %8 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   %9 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %8, align 8
// CHECK-NEXT:   ret %"{{.*}}/runtime/internal/runtime.Slice" %9
// CHECK-NEXT: }

// CHECK-LABEL: define linkonce %"{{.*}}/runtime/internal/runtime.Slice" @"main.(*Slice{{\[\[\]string,string\]}}).Append"(ptr %0, %"{{.*}}/runtime/internal/runtime.Slice" %1){{.*}} {
// CHECK-NEXT: _llgo_0:
// CHECK-NEXT:   %2 = getelementptr inbounds %"main.Slice{{\[\[\]string,string\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   %3 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %2, align 8
// CHECK-NEXT:   %4 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %1, 0
// CHECK-NEXT:   %5 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %1, 1
// CHECK-NEXT:   %6 = call %"{{.*}}/runtime/internal/runtime.Slice" @"{{.*}}/runtime/internal/runtime.SliceAppend"(%"{{.*}}/runtime/internal/runtime.Slice" %3, ptr %4, i64 %5, i64 16)
// CHECK-NEXT:   %7 = getelementptr inbounds %"main.Slice{{\[\[\]string,string\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   store %"{{.*}}/runtime/internal/runtime.Slice" %6, ptr %7, align 8
// CHECK-NEXT:   %8 = getelementptr inbounds %"main.Slice{{\[\[\]string,string\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   %9 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %8, align 8
// CHECK-NEXT:   ret %"{{.*}}/runtime/internal/runtime.Slice" %9
// CHECK-NEXT: }

// CHECK-LABEL: define linkonce %"{{.*}}/runtime/internal/runtime.Slice" @"main.(*Slice{{\[\[\]int,int\]}}).Append2"(ptr %0, %"{{.*}}/runtime/internal/runtime.Slice" %1){{.*}} {
// CHECK-NEXT: _llgo_0:
// CHECK-NEXT:   %2 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   %3 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %2, align 8
// CHECK-NEXT:   %4 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %1, 0
// CHECK-NEXT:   %5 = extractvalue %"{{.*}}/runtime/internal/runtime.Slice" %1, 1
// CHECK-NEXT:   %6 = call %"{{.*}}/runtime/internal/runtime.Slice" @"{{.*}}/runtime/internal/runtime.SliceAppend"(%"{{.*}}/runtime/internal/runtime.Slice" %3, ptr %4, i64 %5, i64 8)
// CHECK-NEXT:   %7 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   store %"{{.*}}/runtime/internal/runtime.Slice" %6, ptr %7, align 8
// CHECK-NEXT:   %8 = getelementptr inbounds %"main.Slice{{\[\[\]int,int\]}}", ptr %0, i32 0, i32 0
// CHECK-NEXT:   %9 = load %"{{.*}}/runtime/internal/runtime.Slice", ptr %8, align 8
// CHECK-NEXT:   ret %"{{.*}}/runtime/internal/runtime.Slice" %9
// CHECK-NEXT: }
