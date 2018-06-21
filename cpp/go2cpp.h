#pragma once
#include <string>
#include <memory>
#include <vector>

typedef signed char int8;
typedef signed short int16;
typedef signed int int32;
typedef signed long long int64;
typedef unsigned char uint8;
typedef unsigned short uint16;
typedef unsigned int uint32;
typedef unsigned long long uint64;
typedef float float32;
typedef double float64;
typedef std::string string;

template <typename T>
inline std::shared_ptr<T> __ptr(T& t) {
    return &t;
}
