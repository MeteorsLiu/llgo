; ModuleID = 'main'
source_filename = "main"

%main.S = type { %"github.com/goplus/llgo/internal/runtime.iface" }
%"github.com/goplus/llgo/internal/runtime.iface" = type { ptr, ptr }
%"github.com/goplus/llgo/internal/runtime.String" = type { ptr, i64 }
%main.impl = type {}
%"github.com/goplus/llgo/internal/runtime.eface" = type { ptr, ptr }
%"github.com/goplus/llgo/internal/runtime.Slice" = type { ptr, i64, i64 }
%"github.com/goplus/llgo/internal/abi.Method" = type { %"github.com/goplus/llgo/internal/runtime.String", ptr, ptr, ptr }
%"github.com/goplus/llgo/internal/abi.Imethod" = type { %"github.com/goplus/llgo/internal/runtime.String", ptr }
%"github.com/goplus/llgo/internal/abi.StructField" = type { %"github.com/goplus/llgo/internal/runtime.String", ptr, i64, %"github.com/goplus/llgo/internal/runtime.String", i1 }

@"main.init$guard" = global i1 false, align 1
@0 = private unnamed_addr constant [3 x i8] c"two", align 1
@__llgo_argc = global i32 0, align 4
@__llgo_argv = global ptr null, align 8
@_llgo_main.impl = linkonce global ptr null, align 8
@"_llgo_struct$n1H8J_3prDN3firMwPxBLVTkE5hJ9Di-AqNvaC9jczw" = linkonce global ptr null, align 8
@1 = private unnamed_addr constant [4 x i8] c"main", align 1
@2 = private unnamed_addr constant [3 x i8] c"one", align 1
@3 = private unnamed_addr constant [8 x i8] c"main.one", align 1
@"_llgo_func$ETeB8WwW04JEq0ztcm-XPTJtuYvtpkjIsAc0-2NT9zA" = linkonce global ptr null, align 8
@_llgo_int = linkonce global ptr null, align 8
@4 = private unnamed_addr constant [8 x i8] c"main.two", align 1
@"_llgo_func$zNDVRsWTIpUPKouNUS805RGX--IV9qVK8B31IZbg5to" = linkonce global ptr null, align 8
@_llgo_string = linkonce global ptr null, align 8
@5 = private unnamed_addr constant [4 x i8] c"impl", align 1
@"main.iface$zZ89tENb5h_KNjvpxf1TXPfaWFYn0IZrZwyVf42lRtA" = linkonce global ptr null, align 8
@_llgo_main.I = linkonce global ptr null, align 8
@6 = private unnamed_addr constant [6 x i8] c"main.I", align 1
@7 = private unnamed_addr constant [21 x i8] c"type assertion failed", align 1
@8 = private unnamed_addr constant [4 x i8] c"pass", align 1

define i64 @main.S.one(%main.S %0) {
_llgo_0:
  %1 = alloca %main.S, align 8
  call void @llvm.memset(ptr %1, i8 0, i64 16, i1 false)
  store %main.S %0, ptr %1, align 8
  %2 = getelementptr inbounds %main.S, ptr %1, i32 0, i32 0
  %3 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %2, align 8
  %4 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %3)
  %5 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %3, 0
  %6 = getelementptr ptr, ptr %5, i64 3
  %7 = load ptr, ptr %6, align 8
  %8 = alloca { ptr, ptr }, align 8
  %9 = getelementptr inbounds { ptr, ptr }, ptr %8, i32 0, i32 0
  store ptr %7, ptr %9, align 8
  %10 = getelementptr inbounds { ptr, ptr }, ptr %8, i32 0, i32 1
  store ptr %4, ptr %10, align 8
  %11 = load { ptr, ptr }, ptr %8, align 8
  %12 = extractvalue { ptr, ptr } %11, 1
  %13 = extractvalue { ptr, ptr } %11, 0
  %14 = call i64 %13(ptr %12)
  ret i64 %14
}

define %"github.com/goplus/llgo/internal/runtime.String" @main.S.two(%main.S %0) {
_llgo_0:
  %1 = alloca %main.S, align 8
  call void @llvm.memset(ptr %1, i8 0, i64 16, i1 false)
  store %main.S %0, ptr %1, align 8
  %2 = getelementptr inbounds %main.S, ptr %1, i32 0, i32 0
  %3 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %2, align 8
  %4 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %3)
  %5 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %3, 0
  %6 = getelementptr ptr, ptr %5, i64 4
  %7 = load ptr, ptr %6, align 8
  %8 = alloca { ptr, ptr }, align 8
  %9 = getelementptr inbounds { ptr, ptr }, ptr %8, i32 0, i32 0
  store ptr %7, ptr %9, align 8
  %10 = getelementptr inbounds { ptr, ptr }, ptr %8, i32 0, i32 1
  store ptr %4, ptr %10, align 8
  %11 = load { ptr, ptr }, ptr %8, align 8
  %12 = extractvalue { ptr, ptr } %11, 1
  %13 = extractvalue { ptr, ptr } %11, 0
  %14 = call %"github.com/goplus/llgo/internal/runtime.String" %13(ptr %12)
  ret %"github.com/goplus/llgo/internal/runtime.String" %14
}

define i64 @"main.(*S).one"(ptr %0) {
_llgo_0:
  %1 = getelementptr inbounds %main.S, ptr %0, i32 0, i32 0
  %2 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %1, align 8
  %3 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %2)
  %4 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %2, 0
  %5 = getelementptr ptr, ptr %4, i64 3
  %6 = load ptr, ptr %5, align 8
  %7 = alloca { ptr, ptr }, align 8
  %8 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 0
  store ptr %6, ptr %8, align 8
  %9 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 1
  store ptr %3, ptr %9, align 8
  %10 = load { ptr, ptr }, ptr %7, align 8
  %11 = extractvalue { ptr, ptr } %10, 1
  %12 = extractvalue { ptr, ptr } %10, 0
  %13 = call i64 %12(ptr %11)
  ret i64 %13
}

define %"github.com/goplus/llgo/internal/runtime.String" @"main.(*S).two"(ptr %0) {
_llgo_0:
  %1 = getelementptr inbounds %main.S, ptr %0, i32 0, i32 0
  %2 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %1, align 8
  %3 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %2)
  %4 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %2, 0
  %5 = getelementptr ptr, ptr %4, i64 4
  %6 = load ptr, ptr %5, align 8
  %7 = alloca { ptr, ptr }, align 8
  %8 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 0
  store ptr %6, ptr %8, align 8
  %9 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 1
  store ptr %3, ptr %9, align 8
  %10 = load { ptr, ptr }, ptr %7, align 8
  %11 = extractvalue { ptr, ptr } %10, 1
  %12 = extractvalue { ptr, ptr } %10, 0
  %13 = call %"github.com/goplus/llgo/internal/runtime.String" %12(ptr %11)
  ret %"github.com/goplus/llgo/internal/runtime.String" %13
}

define i64 @main.impl.one(%main.impl %0) {
_llgo_0:
  ret i64 1
}

define %"github.com/goplus/llgo/internal/runtime.String" @main.impl.two(%main.impl %0) {
_llgo_0:
  %1 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %2 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %1, i32 0, i32 0
  store ptr @0, ptr %2, align 8
  %3 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %1, i32 0, i32 1
  store i64 3, ptr %3, align 4
  %4 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %1, align 8
  ret %"github.com/goplus/llgo/internal/runtime.String" %4
}

define i64 @"main.(*impl).one"(ptr %0) {
_llgo_0:
  %1 = load %main.impl, ptr %0, align 1
  %2 = call i64 @main.impl.one(%main.impl %1)
  ret i64 %2
}

define %"github.com/goplus/llgo/internal/runtime.String" @"main.(*impl).two"(ptr %0) {
_llgo_0:
  %1 = load %main.impl, ptr %0, align 1
  %2 = call %"github.com/goplus/llgo/internal/runtime.String" @main.impl.two(%main.impl %1)
  ret %"github.com/goplus/llgo/internal/runtime.String" %2
}

define void @main.init() {
_llgo_0:
  %0 = load i1, ptr @"main.init$guard", align 1
  br i1 %0, label %_llgo_2, label %_llgo_1

_llgo_1:                                          ; preds = %_llgo_0
  store i1 true, ptr @"main.init$guard", align 1
  call void @"main.init$after"()
  br label %_llgo_2

_llgo_2:                                          ; preds = %_llgo_1, %_llgo_0
  ret void
}

define i32 @main(i32 %0, ptr %1) {
_llgo_0:
  store i32 %0, ptr @__llgo_argc, align 4
  store ptr %1, ptr @__llgo_argv, align 8
  call void @"github.com/goplus/llgo/internal/runtime.init"()
  call void @main.init()
  %2 = alloca %main.S, align 8
  call void @llvm.memset(ptr %2, i8 0, i64 16, i1 false)
  %3 = getelementptr inbounds %main.S, ptr %2, i32 0, i32 0
  %4 = load ptr, ptr @_llgo_main.impl, align 8
  %5 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 0)
  store %main.impl zeroinitializer, ptr %5, align 1
  %6 = load ptr, ptr @"main.iface$zZ89tENb5h_KNjvpxf1TXPfaWFYn0IZrZwyVf42lRtA", align 8
  %7 = call ptr @"github.com/goplus/llgo/internal/runtime.NewItab"(ptr %6, ptr %4)
  %8 = alloca %"github.com/goplus/llgo/internal/runtime.iface", align 8
  %9 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %8, i32 0, i32 0
  store ptr %7, ptr %9, align 8
  %10 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %8, i32 0, i32 1
  store ptr %5, ptr %10, align 8
  %11 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %8, align 8
  store %"github.com/goplus/llgo/internal/runtime.iface" %11, ptr %3, align 8
  %12 = getelementptr inbounds %main.S, ptr %2, i32 0, i32 0
  %13 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %12, align 8
  %14 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %13)
  %15 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %13, 0
  %16 = getelementptr ptr, ptr %15, i64 3
  %17 = load ptr, ptr %16, align 8
  %18 = alloca { ptr, ptr }, align 8
  %19 = getelementptr inbounds { ptr, ptr }, ptr %18, i32 0, i32 0
  store ptr %17, ptr %19, align 8
  %20 = getelementptr inbounds { ptr, ptr }, ptr %18, i32 0, i32 1
  store ptr %14, ptr %20, align 8
  %21 = load { ptr, ptr }, ptr %18, align 8
  %22 = extractvalue { ptr, ptr } %21, 1
  %23 = extractvalue { ptr, ptr } %21, 0
  %24 = call i64 %23(ptr %22)
  %25 = icmp ne i64 %24, 1
  br i1 %25, label %_llgo_1, label %_llgo_2

_llgo_1:                                          ; preds = %_llgo_0
  %26 = load ptr, ptr @_llgo_int, align 8
  %27 = inttoptr i64 %24 to ptr
  %28 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %29 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %28, i32 0, i32 0
  store ptr %26, ptr %29, align 8
  %30 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %28, i32 0, i32 1
  store ptr %27, ptr %30, align 8
  %31 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %28, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %31)
  unreachable

_llgo_2:                                          ; preds = %_llgo_0
  %32 = load %main.S, ptr %2, align 8
  %33 = extractvalue %main.S %32, 0
  %34 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %33)
  %35 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %33, 0
  %36 = getelementptr ptr, ptr %35, i64 3
  %37 = load ptr, ptr %36, align 8
  %38 = alloca { ptr, ptr }, align 8
  %39 = getelementptr inbounds { ptr, ptr }, ptr %38, i32 0, i32 0
  store ptr %37, ptr %39, align 8
  %40 = getelementptr inbounds { ptr, ptr }, ptr %38, i32 0, i32 1
  store ptr %34, ptr %40, align 8
  %41 = load { ptr, ptr }, ptr %38, align 8
  %42 = extractvalue { ptr, ptr } %41, 1
  %43 = extractvalue { ptr, ptr } %41, 0
  %44 = call i64 %43(ptr %42)
  %45 = icmp ne i64 %44, 1
  br i1 %45, label %_llgo_3, label %_llgo_4

_llgo_3:                                          ; preds = %_llgo_2
  %46 = load ptr, ptr @_llgo_int, align 8
  %47 = inttoptr i64 %44 to ptr
  %48 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %49 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %48, i32 0, i32 0
  store ptr %46, ptr %49, align 8
  %50 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %48, i32 0, i32 1
  store ptr %47, ptr %50, align 8
  %51 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %48, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %51)
  unreachable

_llgo_4:                                          ; preds = %_llgo_2
  %52 = getelementptr inbounds %main.S, ptr %2, i32 0, i32 0
  %53 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %52, align 8
  %54 = call ptr @"github.com/goplus/llgo/internal/runtime.IfaceType"(%"github.com/goplus/llgo/internal/runtime.iface" %53)
  %55 = load ptr, ptr @_llgo_main.I, align 8
  %56 = call i1 @"github.com/goplus/llgo/internal/runtime.Implements"(ptr %55, ptr %54)
  br i1 %56, label %_llgo_17, label %_llgo_18

_llgo_5:                                          ; preds = %_llgo_17
  %57 = load ptr, ptr @_llgo_int, align 8
  %58 = inttoptr i64 %165 to ptr
  %59 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %60 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %59, i32 0, i32 0
  store ptr %57, ptr %60, align 8
  %61 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %59, i32 0, i32 1
  store ptr %58, ptr %61, align 8
  %62 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %59, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %62)
  unreachable

_llgo_6:                                          ; preds = %_llgo_17
  %63 = load %main.S, ptr %2, align 8
  %64 = extractvalue %main.S %63, 0
  %65 = call ptr @"github.com/goplus/llgo/internal/runtime.IfaceType"(%"github.com/goplus/llgo/internal/runtime.iface" %64)
  %66 = load ptr, ptr @_llgo_main.I, align 8
  %67 = call i1 @"github.com/goplus/llgo/internal/runtime.Implements"(ptr %66, ptr %65)
  br i1 %67, label %_llgo_19, label %_llgo_20

_llgo_7:                                          ; preds = %_llgo_19
  %68 = load ptr, ptr @_llgo_int, align 8
  %69 = inttoptr i64 %192 to ptr
  %70 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %71 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %70, i32 0, i32 0
  store ptr %68, ptr %71, align 8
  %72 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %70, i32 0, i32 1
  store ptr %69, ptr %72, align 8
  %73 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %70, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %73)
  unreachable

_llgo_8:                                          ; preds = %_llgo_19
  %74 = getelementptr inbounds %main.S, ptr %2, i32 0, i32 0
  %75 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %74, align 8
  %76 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %75)
  %77 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %75, 0
  %78 = getelementptr ptr, ptr %77, i64 4
  %79 = load ptr, ptr %78, align 8
  %80 = alloca { ptr, ptr }, align 8
  %81 = getelementptr inbounds { ptr, ptr }, ptr %80, i32 0, i32 0
  store ptr %79, ptr %81, align 8
  %82 = getelementptr inbounds { ptr, ptr }, ptr %80, i32 0, i32 1
  store ptr %76, ptr %82, align 8
  %83 = load { ptr, ptr }, ptr %80, align 8
  %84 = extractvalue { ptr, ptr } %83, 1
  %85 = extractvalue { ptr, ptr } %83, 0
  %86 = call %"github.com/goplus/llgo/internal/runtime.String" %85(ptr %84)
  %87 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %88 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %87, i32 0, i32 0
  store ptr @0, ptr %88, align 8
  %89 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %87, i32 0, i32 1
  store i64 3, ptr %89, align 4
  %90 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %87, align 8
  %91 = call i1 @"github.com/goplus/llgo/internal/runtime.StringEqual"(%"github.com/goplus/llgo/internal/runtime.String" %86, %"github.com/goplus/llgo/internal/runtime.String" %90)
  %92 = xor i1 %91, true
  br i1 %92, label %_llgo_9, label %_llgo_10

_llgo_9:                                          ; preds = %_llgo_8
  %93 = load ptr, ptr @_llgo_string, align 8
  %94 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %86, ptr %94, align 8
  %95 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %96 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %95, i32 0, i32 0
  store ptr %93, ptr %96, align 8
  %97 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %95, i32 0, i32 1
  store ptr %94, ptr %97, align 8
  %98 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %95, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %98)
  unreachable

_llgo_10:                                         ; preds = %_llgo_8
  %99 = load %main.S, ptr %2, align 8
  %100 = extractvalue %main.S %99, 0
  %101 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %100)
  %102 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %100, 0
  %103 = getelementptr ptr, ptr %102, i64 4
  %104 = load ptr, ptr %103, align 8
  %105 = alloca { ptr, ptr }, align 8
  %106 = getelementptr inbounds { ptr, ptr }, ptr %105, i32 0, i32 0
  store ptr %104, ptr %106, align 8
  %107 = getelementptr inbounds { ptr, ptr }, ptr %105, i32 0, i32 1
  store ptr %101, ptr %107, align 8
  %108 = load { ptr, ptr }, ptr %105, align 8
  %109 = extractvalue { ptr, ptr } %108, 1
  %110 = extractvalue { ptr, ptr } %108, 0
  %111 = call %"github.com/goplus/llgo/internal/runtime.String" %110(ptr %109)
  %112 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %113 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %112, i32 0, i32 0
  store ptr @0, ptr %113, align 8
  %114 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %112, i32 0, i32 1
  store i64 3, ptr %114, align 4
  %115 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %112, align 8
  %116 = call i1 @"github.com/goplus/llgo/internal/runtime.StringEqual"(%"github.com/goplus/llgo/internal/runtime.String" %111, %"github.com/goplus/llgo/internal/runtime.String" %115)
  %117 = xor i1 %116, true
  br i1 %117, label %_llgo_11, label %_llgo_12

_llgo_11:                                         ; preds = %_llgo_10
  %118 = load ptr, ptr @_llgo_string, align 8
  %119 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %111, ptr %119, align 8
  %120 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %121 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %120, i32 0, i32 0
  store ptr %118, ptr %121, align 8
  %122 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %120, i32 0, i32 1
  store ptr %119, ptr %122, align 8
  %123 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %120, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %123)
  unreachable

_llgo_12:                                         ; preds = %_llgo_10
  %124 = getelementptr inbounds %main.S, ptr %2, i32 0, i32 0
  %125 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %124, align 8
  %126 = call ptr @"github.com/goplus/llgo/internal/runtime.IfaceType"(%"github.com/goplus/llgo/internal/runtime.iface" %125)
  %127 = load ptr, ptr @_llgo_main.I, align 8
  %128 = call i1 @"github.com/goplus/llgo/internal/runtime.Implements"(ptr %127, ptr %126)
  br i1 %128, label %_llgo_21, label %_llgo_22

_llgo_13:                                         ; preds = %_llgo_21
  %129 = load ptr, ptr @_llgo_string, align 8
  %130 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %219, ptr %130, align 8
  %131 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %132 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %131, i32 0, i32 0
  store ptr %129, ptr %132, align 8
  %133 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %131, i32 0, i32 1
  store ptr %130, ptr %133, align 8
  %134 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %131, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %134)
  unreachable

_llgo_14:                                         ; preds = %_llgo_21
  %135 = load %main.S, ptr %2, align 8
  %136 = extractvalue %main.S %135, 0
  %137 = call ptr @"github.com/goplus/llgo/internal/runtime.IfaceType"(%"github.com/goplus/llgo/internal/runtime.iface" %136)
  %138 = load ptr, ptr @_llgo_main.I, align 8
  %139 = call i1 @"github.com/goplus/llgo/internal/runtime.Implements"(ptr %138, ptr %137)
  br i1 %139, label %_llgo_23, label %_llgo_24

_llgo_15:                                         ; preds = %_llgo_23
  %140 = load ptr, ptr @_llgo_string, align 8
  %141 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %251, ptr %141, align 8
  %142 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %143 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %142, i32 0, i32 0
  store ptr %140, ptr %143, align 8
  %144 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %142, i32 0, i32 1
  store ptr %141, ptr %144, align 8
  %145 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %142, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %145)
  unreachable

_llgo_16:                                         ; preds = %_llgo_23
  %146 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %147 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %146, i32 0, i32 0
  store ptr @8, ptr %147, align 8
  %148 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %146, i32 0, i32 1
  store i64 4, ptr %148, align 4
  %149 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %146, align 8
  call void @"github.com/goplus/llgo/internal/runtime.PrintString"(%"github.com/goplus/llgo/internal/runtime.String" %149)
  call void @"github.com/goplus/llgo/internal/runtime.PrintByte"(i8 10)
  ret i32 0

_llgo_17:                                         ; preds = %_llgo_4
  %150 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %53, 1
  %151 = load ptr, ptr @"main.iface$zZ89tENb5h_KNjvpxf1TXPfaWFYn0IZrZwyVf42lRtA", align 8
  %152 = call ptr @"github.com/goplus/llgo/internal/runtime.NewItab"(ptr %151, ptr %54)
  %153 = alloca %"github.com/goplus/llgo/internal/runtime.iface", align 8
  %154 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %153, i32 0, i32 0
  store ptr %152, ptr %154, align 8
  %155 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %153, i32 0, i32 1
  store ptr %150, ptr %155, align 8
  %156 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %153, align 8
  %157 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  %158 = getelementptr inbounds { %"github.com/goplus/llgo/internal/runtime.iface" }, ptr %157, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.iface" %53, ptr %158, align 8
  %159 = alloca { ptr, ptr }, align 8
  %160 = getelementptr inbounds { ptr, ptr }, ptr %159, i32 0, i32 0
  store ptr @"main.one$bound", ptr %160, align 8
  %161 = getelementptr inbounds { ptr, ptr }, ptr %159, i32 0, i32 1
  store ptr %157, ptr %161, align 8
  %162 = load { ptr, ptr }, ptr %159, align 8
  %163 = extractvalue { ptr, ptr } %162, 1
  %164 = extractvalue { ptr, ptr } %162, 0
  %165 = call i64 %164(ptr %163)
  %166 = icmp ne i64 %165, 1
  br i1 %166, label %_llgo_5, label %_llgo_6

_llgo_18:                                         ; preds = %_llgo_4
  %167 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %168 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %167, i32 0, i32 0
  store ptr @7, ptr %168, align 8
  %169 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %167, i32 0, i32 1
  store i64 21, ptr %169, align 4
  %170 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %167, align 8
  %171 = load ptr, ptr @_llgo_string, align 8
  %172 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %170, ptr %172, align 8
  %173 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %174 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %173, i32 0, i32 0
  store ptr %171, ptr %174, align 8
  %175 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %173, i32 0, i32 1
  store ptr %172, ptr %175, align 8
  %176 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %173, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %176)
  unreachable

_llgo_19:                                         ; preds = %_llgo_6
  %177 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %64, 1
  %178 = load ptr, ptr @"main.iface$zZ89tENb5h_KNjvpxf1TXPfaWFYn0IZrZwyVf42lRtA", align 8
  %179 = call ptr @"github.com/goplus/llgo/internal/runtime.NewItab"(ptr %178, ptr %65)
  %180 = alloca %"github.com/goplus/llgo/internal/runtime.iface", align 8
  %181 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %180, i32 0, i32 0
  store ptr %179, ptr %181, align 8
  %182 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %180, i32 0, i32 1
  store ptr %177, ptr %182, align 8
  %183 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %180, align 8
  %184 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  %185 = getelementptr inbounds { %"github.com/goplus/llgo/internal/runtime.iface" }, ptr %184, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.iface" %64, ptr %185, align 8
  %186 = alloca { ptr, ptr }, align 8
  %187 = getelementptr inbounds { ptr, ptr }, ptr %186, i32 0, i32 0
  store ptr @"main.one$bound", ptr %187, align 8
  %188 = getelementptr inbounds { ptr, ptr }, ptr %186, i32 0, i32 1
  store ptr %184, ptr %188, align 8
  %189 = load { ptr, ptr }, ptr %186, align 8
  %190 = extractvalue { ptr, ptr } %189, 1
  %191 = extractvalue { ptr, ptr } %189, 0
  %192 = call i64 %191(ptr %190)
  %193 = icmp ne i64 %192, 1
  br i1 %193, label %_llgo_7, label %_llgo_8

_llgo_20:                                         ; preds = %_llgo_6
  %194 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %195 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %194, i32 0, i32 0
  store ptr @7, ptr %195, align 8
  %196 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %194, i32 0, i32 1
  store i64 21, ptr %196, align 4
  %197 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %194, align 8
  %198 = load ptr, ptr @_llgo_string, align 8
  %199 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %197, ptr %199, align 8
  %200 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %201 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %200, i32 0, i32 0
  store ptr %198, ptr %201, align 8
  %202 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %200, i32 0, i32 1
  store ptr %199, ptr %202, align 8
  %203 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %200, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %203)
  unreachable

_llgo_21:                                         ; preds = %_llgo_12
  %204 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %125, 1
  %205 = load ptr, ptr @"main.iface$zZ89tENb5h_KNjvpxf1TXPfaWFYn0IZrZwyVf42lRtA", align 8
  %206 = call ptr @"github.com/goplus/llgo/internal/runtime.NewItab"(ptr %205, ptr %126)
  %207 = alloca %"github.com/goplus/llgo/internal/runtime.iface", align 8
  %208 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %207, i32 0, i32 0
  store ptr %206, ptr %208, align 8
  %209 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %207, i32 0, i32 1
  store ptr %204, ptr %209, align 8
  %210 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %207, align 8
  %211 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  %212 = getelementptr inbounds { %"github.com/goplus/llgo/internal/runtime.iface" }, ptr %211, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.iface" %125, ptr %212, align 8
  %213 = alloca { ptr, ptr }, align 8
  %214 = getelementptr inbounds { ptr, ptr }, ptr %213, i32 0, i32 0
  store ptr @"main.two$bound", ptr %214, align 8
  %215 = getelementptr inbounds { ptr, ptr }, ptr %213, i32 0, i32 1
  store ptr %211, ptr %215, align 8
  %216 = load { ptr, ptr }, ptr %213, align 8
  %217 = extractvalue { ptr, ptr } %216, 1
  %218 = extractvalue { ptr, ptr } %216, 0
  %219 = call %"github.com/goplus/llgo/internal/runtime.String" %218(ptr %217)
  %220 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %221 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %220, i32 0, i32 0
  store ptr @0, ptr %221, align 8
  %222 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %220, i32 0, i32 1
  store i64 3, ptr %222, align 4
  %223 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %220, align 8
  %224 = call i1 @"github.com/goplus/llgo/internal/runtime.StringEqual"(%"github.com/goplus/llgo/internal/runtime.String" %219, %"github.com/goplus/llgo/internal/runtime.String" %223)
  %225 = xor i1 %224, true
  br i1 %225, label %_llgo_13, label %_llgo_14

_llgo_22:                                         ; preds = %_llgo_12
  %226 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %227 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %226, i32 0, i32 0
  store ptr @7, ptr %227, align 8
  %228 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %226, i32 0, i32 1
  store i64 21, ptr %228, align 4
  %229 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %226, align 8
  %230 = load ptr, ptr @_llgo_string, align 8
  %231 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %229, ptr %231, align 8
  %232 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %233 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %232, i32 0, i32 0
  store ptr %230, ptr %233, align 8
  %234 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %232, i32 0, i32 1
  store ptr %231, ptr %234, align 8
  %235 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %232, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %235)
  unreachable

_llgo_23:                                         ; preds = %_llgo_14
  %236 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %136, 1
  %237 = load ptr, ptr @"main.iface$zZ89tENb5h_KNjvpxf1TXPfaWFYn0IZrZwyVf42lRtA", align 8
  %238 = call ptr @"github.com/goplus/llgo/internal/runtime.NewItab"(ptr %237, ptr %137)
  %239 = alloca %"github.com/goplus/llgo/internal/runtime.iface", align 8
  %240 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %239, i32 0, i32 0
  store ptr %238, ptr %240, align 8
  %241 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.iface", ptr %239, i32 0, i32 1
  store ptr %236, ptr %241, align 8
  %242 = load %"github.com/goplus/llgo/internal/runtime.iface", ptr %239, align 8
  %243 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  %244 = getelementptr inbounds { %"github.com/goplus/llgo/internal/runtime.iface" }, ptr %243, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.iface" %136, ptr %244, align 8
  %245 = alloca { ptr, ptr }, align 8
  %246 = getelementptr inbounds { ptr, ptr }, ptr %245, i32 0, i32 0
  store ptr @"main.two$bound", ptr %246, align 8
  %247 = getelementptr inbounds { ptr, ptr }, ptr %245, i32 0, i32 1
  store ptr %243, ptr %247, align 8
  %248 = load { ptr, ptr }, ptr %245, align 8
  %249 = extractvalue { ptr, ptr } %248, 1
  %250 = extractvalue { ptr, ptr } %248, 0
  %251 = call %"github.com/goplus/llgo/internal/runtime.String" %250(ptr %249)
  %252 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %253 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %252, i32 0, i32 0
  store ptr @0, ptr %253, align 8
  %254 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %252, i32 0, i32 1
  store i64 3, ptr %254, align 4
  %255 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %252, align 8
  %256 = call i1 @"github.com/goplus/llgo/internal/runtime.StringEqual"(%"github.com/goplus/llgo/internal/runtime.String" %251, %"github.com/goplus/llgo/internal/runtime.String" %255)
  %257 = xor i1 %256, true
  br i1 %257, label %_llgo_15, label %_llgo_16

_llgo_24:                                         ; preds = %_llgo_14
  %258 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %259 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %258, i32 0, i32 0
  store ptr @7, ptr %259, align 8
  %260 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %258, i32 0, i32 1
  store i64 21, ptr %260, align 4
  %261 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %258, align 8
  %262 = load ptr, ptr @_llgo_string, align 8
  %263 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 16)
  store %"github.com/goplus/llgo/internal/runtime.String" %261, ptr %263, align 8
  %264 = alloca %"github.com/goplus/llgo/internal/runtime.eface", align 8
  %265 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %264, i32 0, i32 0
  store ptr %262, ptr %265, align 8
  %266 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.eface", ptr %264, i32 0, i32 1
  store ptr %263, ptr %266, align 8
  %267 = load %"github.com/goplus/llgo/internal/runtime.eface", ptr %264, align 8
  call void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface" %267)
  unreachable
}

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: write)
declare void @llvm.memset(ptr nocapture writeonly, i8, i64, i1 immarg) #0

declare ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface")

declare void @"github.com/goplus/llgo/internal/runtime.init"()

define void @"main.init$after"() {
_llgo_0:
  %0 = call ptr @"github.com/goplus/llgo/internal/runtime.NewNamed"(i64 25, i64 0, i64 2, i64 2)
  store ptr %0, ptr @_llgo_main.impl, align 8
  %1 = load ptr, ptr @"_llgo_struct$n1H8J_3prDN3firMwPxBLVTkE5hJ9Di-AqNvaC9jczw", align 8
  %2 = icmp eq ptr %1, null
  br i1 %2, label %_llgo_1, label %_llgo_2

_llgo_1:                                          ; preds = %_llgo_0
  %3 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %4 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %3, i32 0, i32 0
  store ptr @1, ptr %4, align 8
  %5 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %3, i32 0, i32 1
  store i64 4, ptr %5, align 4
  %6 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %3, align 8
  %7 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 0)
  %8 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %9 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %8, i32 0, i32 0
  store ptr %7, ptr %9, align 8
  %10 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %8, i32 0, i32 1
  store i64 0, ptr %10, align 4
  %11 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %8, i32 0, i32 2
  store i64 0, ptr %11, align 4
  %12 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %8, align 8
  %13 = call ptr @"github.com/goplus/llgo/internal/runtime.Struct"(%"github.com/goplus/llgo/internal/runtime.String" %6, i64 0, %"github.com/goplus/llgo/internal/runtime.Slice" %12)
  store ptr %13, ptr @"_llgo_struct$n1H8J_3prDN3firMwPxBLVTkE5hJ9Di-AqNvaC9jczw", align 8
  br label %_llgo_2

_llgo_2:                                          ; preds = %_llgo_1, %_llgo_0
  %14 = load ptr, ptr @"_llgo_struct$n1H8J_3prDN3firMwPxBLVTkE5hJ9Di-AqNvaC9jczw", align 8
  %15 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %16 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %15, i32 0, i32 0
  store ptr @2, ptr %16, align 8
  %17 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %15, i32 0, i32 1
  store i64 3, ptr %17, align 4
  %18 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %15, align 8
  %19 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %20 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %19, i32 0, i32 0
  store ptr @3, ptr %20, align 8
  %21 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %19, i32 0, i32 1
  store i64 8, ptr %21, align 4
  %22 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %19, align 8
  %23 = load ptr, ptr @_llgo_int, align 8
  %24 = icmp eq ptr %23, null
  br i1 %24, label %_llgo_3, label %_llgo_4

_llgo_3:                                          ; preds = %_llgo_2
  %25 = call ptr @"github.com/goplus/llgo/internal/runtime.Basic"(i64 34)
  store ptr %25, ptr @_llgo_int, align 8
  br label %_llgo_4

_llgo_4:                                          ; preds = %_llgo_3, %_llgo_2
  %26 = load ptr, ptr @_llgo_int, align 8
  %27 = load ptr, ptr @"_llgo_func$ETeB8WwW04JEq0ztcm-XPTJtuYvtpkjIsAc0-2NT9zA", align 8
  %28 = icmp eq ptr %27, null
  br i1 %28, label %_llgo_5, label %_llgo_6

_llgo_5:                                          ; preds = %_llgo_4
  %29 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 0)
  %30 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %31 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %30, i32 0, i32 0
  store ptr %29, ptr %31, align 8
  %32 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %30, i32 0, i32 1
  store i64 0, ptr %32, align 4
  %33 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %30, i32 0, i32 2
  store i64 0, ptr %33, align 4
  %34 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %30, align 8
  %35 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 8)
  %36 = getelementptr ptr, ptr %35, i64 0
  store ptr %26, ptr %36, align 8
  %37 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %38 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %37, i32 0, i32 0
  store ptr %35, ptr %38, align 8
  %39 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %37, i32 0, i32 1
  store i64 1, ptr %39, align 4
  %40 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %37, i32 0, i32 2
  store i64 1, ptr %40, align 4
  %41 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %37, align 8
  %42 = call ptr @"github.com/goplus/llgo/internal/runtime.Func"(%"github.com/goplus/llgo/internal/runtime.Slice" %34, %"github.com/goplus/llgo/internal/runtime.Slice" %41, i1 false)
  call void @"github.com/goplus/llgo/internal/runtime.SetDirectIface"(ptr %42)
  store ptr %42, ptr @"_llgo_func$ETeB8WwW04JEq0ztcm-XPTJtuYvtpkjIsAc0-2NT9zA", align 8
  br label %_llgo_6

_llgo_6:                                          ; preds = %_llgo_5, %_llgo_4
  %43 = load ptr, ptr @"_llgo_func$ETeB8WwW04JEq0ztcm-XPTJtuYvtpkjIsAc0-2NT9zA", align 8
  %44 = alloca %"github.com/goplus/llgo/internal/abi.Method", align 8
  %45 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %44, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %22, ptr %45, align 8
  %46 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %44, i32 0, i32 1
  store ptr %43, ptr %46, align 8
  %47 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %44, i32 0, i32 2
  store ptr @"main.(*impl).one", ptr %47, align 8
  %48 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %44, i32 0, i32 3
  store ptr @"main.(*impl).one", ptr %48, align 8
  %49 = load %"github.com/goplus/llgo/internal/abi.Method", ptr %44, align 8
  %50 = alloca %"github.com/goplus/llgo/internal/abi.Method", align 8
  %51 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %50, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %22, ptr %51, align 8
  %52 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %50, i32 0, i32 1
  store ptr %43, ptr %52, align 8
  %53 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %50, i32 0, i32 2
  store ptr @"main.(*impl).one", ptr %53, align 8
  %54 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %50, i32 0, i32 3
  store ptr @main.impl.one, ptr %54, align 8
  %55 = load %"github.com/goplus/llgo/internal/abi.Method", ptr %50, align 8
  %56 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %57 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %56, i32 0, i32 0
  store ptr @0, ptr %57, align 8
  %58 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %56, i32 0, i32 1
  store i64 3, ptr %58, align 4
  %59 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %56, align 8
  %60 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %61 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %60, i32 0, i32 0
  store ptr @4, ptr %61, align 8
  %62 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %60, i32 0, i32 1
  store i64 8, ptr %62, align 4
  %63 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %60, align 8
  %64 = load ptr, ptr @_llgo_string, align 8
  %65 = icmp eq ptr %64, null
  br i1 %65, label %_llgo_7, label %_llgo_8

_llgo_7:                                          ; preds = %_llgo_6
  %66 = call ptr @"github.com/goplus/llgo/internal/runtime.Basic"(i64 24)
  store ptr %66, ptr @_llgo_string, align 8
  br label %_llgo_8

_llgo_8:                                          ; preds = %_llgo_7, %_llgo_6
  %67 = load ptr, ptr @_llgo_string, align 8
  %68 = load ptr, ptr @"_llgo_func$zNDVRsWTIpUPKouNUS805RGX--IV9qVK8B31IZbg5to", align 8
  %69 = icmp eq ptr %68, null
  br i1 %69, label %_llgo_9, label %_llgo_10

_llgo_9:                                          ; preds = %_llgo_8
  %70 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 0)
  %71 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %72 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %71, i32 0, i32 0
  store ptr %70, ptr %72, align 8
  %73 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %71, i32 0, i32 1
  store i64 0, ptr %73, align 4
  %74 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %71, i32 0, i32 2
  store i64 0, ptr %74, align 4
  %75 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %71, align 8
  %76 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 8)
  %77 = getelementptr ptr, ptr %76, i64 0
  store ptr %67, ptr %77, align 8
  %78 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %79 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %78, i32 0, i32 0
  store ptr %76, ptr %79, align 8
  %80 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %78, i32 0, i32 1
  store i64 1, ptr %80, align 4
  %81 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %78, i32 0, i32 2
  store i64 1, ptr %81, align 4
  %82 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %78, align 8
  %83 = call ptr @"github.com/goplus/llgo/internal/runtime.Func"(%"github.com/goplus/llgo/internal/runtime.Slice" %75, %"github.com/goplus/llgo/internal/runtime.Slice" %82, i1 false)
  call void @"github.com/goplus/llgo/internal/runtime.SetDirectIface"(ptr %83)
  store ptr %83, ptr @"_llgo_func$zNDVRsWTIpUPKouNUS805RGX--IV9qVK8B31IZbg5to", align 8
  br label %_llgo_10

_llgo_10:                                         ; preds = %_llgo_9, %_llgo_8
  %84 = load ptr, ptr @"_llgo_func$zNDVRsWTIpUPKouNUS805RGX--IV9qVK8B31IZbg5to", align 8
  %85 = alloca %"github.com/goplus/llgo/internal/abi.Method", align 8
  %86 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %85, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %63, ptr %86, align 8
  %87 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %85, i32 0, i32 1
  store ptr %84, ptr %87, align 8
  %88 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %85, i32 0, i32 2
  store ptr @"main.(*impl).two", ptr %88, align 8
  %89 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %85, i32 0, i32 3
  store ptr @"main.(*impl).two", ptr %89, align 8
  %90 = load %"github.com/goplus/llgo/internal/abi.Method", ptr %85, align 8
  %91 = alloca %"github.com/goplus/llgo/internal/abi.Method", align 8
  %92 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %91, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %63, ptr %92, align 8
  %93 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %91, i32 0, i32 1
  store ptr %84, ptr %93, align 8
  %94 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %91, i32 0, i32 2
  store ptr @"main.(*impl).two", ptr %94, align 8
  %95 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Method", ptr %91, i32 0, i32 3
  store ptr @main.impl.two, ptr %95, align 8
  %96 = load %"github.com/goplus/llgo/internal/abi.Method", ptr %91, align 8
  %97 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 80)
  %98 = getelementptr %"github.com/goplus/llgo/internal/abi.Method", ptr %97, i64 0
  store %"github.com/goplus/llgo/internal/abi.Method" %55, ptr %98, align 8
  %99 = getelementptr %"github.com/goplus/llgo/internal/abi.Method", ptr %97, i64 1
  store %"github.com/goplus/llgo/internal/abi.Method" %96, ptr %99, align 8
  %100 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %101 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %100, i32 0, i32 0
  store ptr %97, ptr %101, align 8
  %102 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %100, i32 0, i32 1
  store i64 2, ptr %102, align 4
  %103 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %100, i32 0, i32 2
  store i64 2, ptr %103, align 4
  %104 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %100, align 8
  %105 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 80)
  %106 = getelementptr %"github.com/goplus/llgo/internal/abi.Method", ptr %105, i64 0
  store %"github.com/goplus/llgo/internal/abi.Method" %49, ptr %106, align 8
  %107 = getelementptr %"github.com/goplus/llgo/internal/abi.Method", ptr %105, i64 1
  store %"github.com/goplus/llgo/internal/abi.Method" %90, ptr %107, align 8
  %108 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %109 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %108, i32 0, i32 0
  store ptr %105, ptr %109, align 8
  %110 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %108, i32 0, i32 1
  store i64 2, ptr %110, align 4
  %111 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %108, i32 0, i32 2
  store i64 2, ptr %111, align 4
  %112 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %108, align 8
  %113 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %114 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %113, i32 0, i32 0
  store ptr @1, ptr %114, align 8
  %115 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %113, i32 0, i32 1
  store i64 4, ptr %115, align 4
  %116 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %113, align 8
  %117 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %118 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %117, i32 0, i32 0
  store ptr @5, ptr %118, align 8
  %119 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %117, i32 0, i32 1
  store i64 4, ptr %119, align 4
  %120 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %117, align 8
  call void @"github.com/goplus/llgo/internal/runtime.InitNamed"(ptr %0, %"github.com/goplus/llgo/internal/runtime.String" %116, %"github.com/goplus/llgo/internal/runtime.String" %120, ptr %14, %"github.com/goplus/llgo/internal/runtime.Slice" %104, %"github.com/goplus/llgo/internal/runtime.Slice" %112)
  %121 = load ptr, ptr @"_llgo_func$ETeB8WwW04JEq0ztcm-XPTJtuYvtpkjIsAc0-2NT9zA", align 8
  %122 = load ptr, ptr @"_llgo_func$zNDVRsWTIpUPKouNUS805RGX--IV9qVK8B31IZbg5to", align 8
  %123 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %124 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %123, i32 0, i32 0
  store ptr @3, ptr %124, align 8
  %125 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %123, i32 0, i32 1
  store i64 8, ptr %125, align 4
  %126 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %123, align 8
  %127 = alloca %"github.com/goplus/llgo/internal/abi.Imethod", align 8
  %128 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %127, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %126, ptr %128, align 8
  %129 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %127, i32 0, i32 1
  store ptr %121, ptr %129, align 8
  %130 = load %"github.com/goplus/llgo/internal/abi.Imethod", ptr %127, align 8
  %131 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %132 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %131, i32 0, i32 0
  store ptr @4, ptr %132, align 8
  %133 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %131, i32 0, i32 1
  store i64 8, ptr %133, align 4
  %134 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %131, align 8
  %135 = alloca %"github.com/goplus/llgo/internal/abi.Imethod", align 8
  %136 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %135, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %134, ptr %136, align 8
  %137 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %135, i32 0, i32 1
  store ptr %122, ptr %137, align 8
  %138 = load %"github.com/goplus/llgo/internal/abi.Imethod", ptr %135, align 8
  %139 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 48)
  %140 = getelementptr %"github.com/goplus/llgo/internal/abi.Imethod", ptr %139, i64 0
  store %"github.com/goplus/llgo/internal/abi.Imethod" %130, ptr %140, align 8
  %141 = getelementptr %"github.com/goplus/llgo/internal/abi.Imethod", ptr %139, i64 1
  store %"github.com/goplus/llgo/internal/abi.Imethod" %138, ptr %141, align 8
  %142 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %143 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %142, i32 0, i32 0
  store ptr %139, ptr %143, align 8
  %144 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %142, i32 0, i32 1
  store i64 2, ptr %144, align 4
  %145 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %142, i32 0, i32 2
  store i64 2, ptr %145, align 4
  %146 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %142, align 8
  %147 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %148 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %147, i32 0, i32 0
  store ptr @1, ptr %148, align 8
  %149 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %147, i32 0, i32 1
  store i64 4, ptr %149, align 4
  %150 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %147, align 8
  %151 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %152 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %151, i32 0, i32 0
  store ptr null, ptr %152, align 8
  %153 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %151, i32 0, i32 1
  store i64 0, ptr %153, align 4
  %154 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %151, align 8
  %155 = call ptr @"github.com/goplus/llgo/internal/runtime.Interface"(%"github.com/goplus/llgo/internal/runtime.String" %150, %"github.com/goplus/llgo/internal/runtime.String" %154, %"github.com/goplus/llgo/internal/runtime.Slice" %146)
  store ptr %155, ptr @"main.iface$zZ89tENb5h_KNjvpxf1TXPfaWFYn0IZrZwyVf42lRtA", align 8
  %156 = load ptr, ptr @"_llgo_func$ETeB8WwW04JEq0ztcm-XPTJtuYvtpkjIsAc0-2NT9zA", align 8
  %157 = load ptr, ptr @"_llgo_func$zNDVRsWTIpUPKouNUS805RGX--IV9qVK8B31IZbg5to", align 8
  %158 = load ptr, ptr @_llgo_main.I, align 8
  %159 = icmp eq ptr %158, null
  br i1 %159, label %_llgo_11, label %_llgo_12

_llgo_11:                                         ; preds = %_llgo_10
  %160 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %161 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %160, i32 0, i32 0
  store ptr @3, ptr %161, align 8
  %162 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %160, i32 0, i32 1
  store i64 8, ptr %162, align 4
  %163 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %160, align 8
  %164 = alloca %"github.com/goplus/llgo/internal/abi.Imethod", align 8
  %165 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %164, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %163, ptr %165, align 8
  %166 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %164, i32 0, i32 1
  store ptr %156, ptr %166, align 8
  %167 = load %"github.com/goplus/llgo/internal/abi.Imethod", ptr %164, align 8
  %168 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %169 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %168, i32 0, i32 0
  store ptr @4, ptr %169, align 8
  %170 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %168, i32 0, i32 1
  store i64 8, ptr %170, align 4
  %171 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %168, align 8
  %172 = alloca %"github.com/goplus/llgo/internal/abi.Imethod", align 8
  %173 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %172, i32 0, i32 0
  store %"github.com/goplus/llgo/internal/runtime.String" %171, ptr %173, align 8
  %174 = getelementptr inbounds %"github.com/goplus/llgo/internal/abi.Imethod", ptr %172, i32 0, i32 1
  store ptr %157, ptr %174, align 8
  %175 = load %"github.com/goplus/llgo/internal/abi.Imethod", ptr %172, align 8
  %176 = call ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64 48)
  %177 = getelementptr %"github.com/goplus/llgo/internal/abi.Imethod", ptr %176, i64 0
  store %"github.com/goplus/llgo/internal/abi.Imethod" %167, ptr %177, align 8
  %178 = getelementptr %"github.com/goplus/llgo/internal/abi.Imethod", ptr %176, i64 1
  store %"github.com/goplus/llgo/internal/abi.Imethod" %175, ptr %178, align 8
  %179 = alloca %"github.com/goplus/llgo/internal/runtime.Slice", align 8
  %180 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %179, i32 0, i32 0
  store ptr %176, ptr %180, align 8
  %181 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %179, i32 0, i32 1
  store i64 2, ptr %181, align 4
  %182 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.Slice", ptr %179, i32 0, i32 2
  store i64 2, ptr %182, align 4
  %183 = load %"github.com/goplus/llgo/internal/runtime.Slice", ptr %179, align 8
  %184 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %185 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %184, i32 0, i32 0
  store ptr @1, ptr %185, align 8
  %186 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %184, i32 0, i32 1
  store i64 4, ptr %186, align 4
  %187 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %184, align 8
  %188 = alloca %"github.com/goplus/llgo/internal/runtime.String", align 8
  %189 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %188, i32 0, i32 0
  store ptr @6, ptr %189, align 8
  %190 = getelementptr inbounds %"github.com/goplus/llgo/internal/runtime.String", ptr %188, i32 0, i32 1
  store i64 6, ptr %190, align 4
  %191 = load %"github.com/goplus/llgo/internal/runtime.String", ptr %188, align 8
  %192 = call ptr @"github.com/goplus/llgo/internal/runtime.Interface"(%"github.com/goplus/llgo/internal/runtime.String" %187, %"github.com/goplus/llgo/internal/runtime.String" %191, %"github.com/goplus/llgo/internal/runtime.Slice" %183)
  store ptr %192, ptr @_llgo_main.I, align 8
  br label %_llgo_12

_llgo_12:                                         ; preds = %_llgo_11, %_llgo_10
  ret void
}

declare ptr @"github.com/goplus/llgo/internal/runtime.NewNamed"(i64, i64, i64, i64)

declare ptr @"github.com/goplus/llgo/internal/runtime.Struct"(%"github.com/goplus/llgo/internal/runtime.String", i64, %"github.com/goplus/llgo/internal/runtime.Slice")

declare %"github.com/goplus/llgo/internal/abi.StructField" @"github.com/goplus/llgo/internal/runtime.StructField"(%"github.com/goplus/llgo/internal/runtime.String", ptr, i64, %"github.com/goplus/llgo/internal/runtime.String", i1)

declare ptr @"github.com/goplus/llgo/internal/runtime.AllocU"(i64)

declare void @"github.com/goplus/llgo/internal/runtime.InitNamed"(ptr, %"github.com/goplus/llgo/internal/runtime.String", %"github.com/goplus/llgo/internal/runtime.String", ptr, %"github.com/goplus/llgo/internal/runtime.Slice", %"github.com/goplus/llgo/internal/runtime.Slice")

declare ptr @"github.com/goplus/llgo/internal/runtime.Basic"(i64)

declare ptr @"github.com/goplus/llgo/internal/runtime.Func"(%"github.com/goplus/llgo/internal/runtime.Slice", %"github.com/goplus/llgo/internal/runtime.Slice", i1)

declare void @"github.com/goplus/llgo/internal/runtime.SetDirectIface"(ptr)

declare ptr @"github.com/goplus/llgo/internal/runtime.Interface"(%"github.com/goplus/llgo/internal/runtime.String", %"github.com/goplus/llgo/internal/runtime.String", %"github.com/goplus/llgo/internal/runtime.Slice")

declare ptr @"github.com/goplus/llgo/internal/runtime.NewItab"(ptr, ptr)

declare void @"github.com/goplus/llgo/internal/runtime.Panic"(%"github.com/goplus/llgo/internal/runtime.eface")

declare ptr @"github.com/goplus/llgo/internal/runtime.IfaceType"(%"github.com/goplus/llgo/internal/runtime.iface")

declare i1 @"github.com/goplus/llgo/internal/runtime.Implements"(ptr, ptr)

define i64 @"main.one$bound"(ptr %0) {
_llgo_0:
  %1 = load { %"github.com/goplus/llgo/internal/runtime.iface" }, ptr %0, align 8
  %2 = extractvalue { %"github.com/goplus/llgo/internal/runtime.iface" } %1, 0
  %3 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %2)
  %4 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %2, 0
  %5 = getelementptr ptr, ptr %4, i64 3
  %6 = load ptr, ptr %5, align 8
  %7 = alloca { ptr, ptr }, align 8
  %8 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 0
  store ptr %6, ptr %8, align 8
  %9 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 1
  store ptr %3, ptr %9, align 8
  %10 = load { ptr, ptr }, ptr %7, align 8
  %11 = extractvalue { ptr, ptr } %10, 1
  %12 = extractvalue { ptr, ptr } %10, 0
  %13 = call i64 %12(ptr %11)
  ret i64 %13
}

declare i1 @"github.com/goplus/llgo/internal/runtime.StringEqual"(%"github.com/goplus/llgo/internal/runtime.String", %"github.com/goplus/llgo/internal/runtime.String")

define %"github.com/goplus/llgo/internal/runtime.String" @"main.two$bound"(ptr %0) {
_llgo_0:
  %1 = load { %"github.com/goplus/llgo/internal/runtime.iface" }, ptr %0, align 8
  %2 = extractvalue { %"github.com/goplus/llgo/internal/runtime.iface" } %1, 0
  %3 = call ptr @"github.com/goplus/llgo/internal/runtime.IfacePtrData"(%"github.com/goplus/llgo/internal/runtime.iface" %2)
  %4 = extractvalue %"github.com/goplus/llgo/internal/runtime.iface" %2, 0
  %5 = getelementptr ptr, ptr %4, i64 4
  %6 = load ptr, ptr %5, align 8
  %7 = alloca { ptr, ptr }, align 8
  %8 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 0
  store ptr %6, ptr %8, align 8
  %9 = getelementptr inbounds { ptr, ptr }, ptr %7, i32 0, i32 1
  store ptr %3, ptr %9, align 8
  %10 = load { ptr, ptr }, ptr %7, align 8
  %11 = extractvalue { ptr, ptr } %10, 1
  %12 = extractvalue { ptr, ptr } %10, 0
  %13 = call %"github.com/goplus/llgo/internal/runtime.String" %12(ptr %11)
  ret %"github.com/goplus/llgo/internal/runtime.String" %13
}

declare void @"github.com/goplus/llgo/internal/runtime.PrintString"(%"github.com/goplus/llgo/internal/runtime.String")

declare void @"github.com/goplus/llgo/internal/runtime.PrintByte"(i8)

attributes #0 = { nocallback nofree nounwind willreturn memory(argmem: write) }
