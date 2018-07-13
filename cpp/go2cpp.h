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
typedef unsigned long long uint;
typedef float float32;
typedef double float64;
typedef std::string string;

template <typename T>
inline std::shared_ptr<T> __ptr(const T& t) {
    return &t;
}

template <typename T>
inline std::vector<T> make(const T* t, int n) {
    std::vector<T> result;
    result.resize(n);
    return result;
}

template <typename T>
inline std::vector<T> append(std::vector<T> s, ...) {
    // va_list arg;
    // va_start(arg, argnum);
    // for ()
    return s;
}

template <typename T>
inline int len(std::vector<T> s) {
    return s.size();
}

inline void defer(void(*fn)()) {
}
