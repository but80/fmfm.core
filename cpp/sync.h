#pragma once
namespace sync {

    struct Mutex {
    };

    void Mutex__Lock(std::shared_ptr<Mutex>);
    void Mutex__Unlock(std::shared_ptr<Mutex>);

}
