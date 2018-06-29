#pragma once
namespace sort {

    template <typename T>
    void Slice(std::vector<T> v, bool (*less)(int, int));

}
