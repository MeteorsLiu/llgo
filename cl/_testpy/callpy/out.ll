; ModuleID = 'main'
source_filename = "main"

@"main.init$guard" = global ptr null
@__llgo_argc = global ptr null
@__llgo_argv = global ptr null
@__llgo_py.math.sqrt = linkonce global ptr null
@0 = private unnamed_addr constant [14 x i8] c"sqrt(2) = %f\0A\00", align 1
@__llgo_py.math = external global ptr
@1 = private unnamed_addr constant [5 x i8] c"sqrt\00", align 1

define void @main.init() {
_llgo_0:
  %0 = load i1, ptr @"main.init$guard", align 1
  br i1 %0, label %_llgo_2, label %_llgo_1

_llgo_1:                                          ; preds = %_llgo_0
  store i1 true, ptr @"main.init$guard", align 1
  call void @"github.com/goplus/llgo/py/math.init"()
  %1 = load ptr, ptr @__llgo_py.math, align 8
  call void (ptr, ...) @llgoLoadPyModSyms(ptr %1, ptr @1, ptr @__llgo_py.math.sqrt, ptr null)
  br label %_llgo_2

_llgo_2:                                          ; preds = %_llgo_1, %_llgo_0
  ret void
}

define void @main(i32 %0, ptr %1) {
_llgo_0:
  call void @Py_Initialize()
  store i32 %0, ptr @__llgo_argc, align 4
  store ptr %1, ptr @__llgo_argv, align 8
  call void @"github.com/goplus/llgo/internal/runtime.init"()
  call void @main.init()
  %2 = call ptr @PyFloat_FromDouble(double 2.000000e+00)
  %3 = load ptr, ptr @__llgo_py.math.sqrt, align 8
  %4 = call ptr @PyObject_CallOneArg(ptr %3, ptr %2)
  %5 = call double @PyFloat_AsDouble(ptr %4)
  %6 = call i32 (ptr, ...) @printf(ptr @0, double %5)
  ret void
}

declare void @"github.com/goplus/llgo/py/math.init"()

declare void @"github.com/goplus/llgo/internal/runtime.init"()

declare ptr @PyFloat_FromDouble(double)

declare ptr @PyObject_CallOneArg(ptr, ptr)

declare double @PyFloat_AsDouble(ptr)

declare i32 @printf(ptr, ...)

declare void @llgoLoadPyModSyms(ptr, ...)

declare void @Py_Initialize()
