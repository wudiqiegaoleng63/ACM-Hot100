import sys

def main():
    data = sys.stdin.read().split()
    if not data:
        return
    n = int(data[0])
    target = int(data[1])
    a = list(map(int, data[2:2 + n]))

    mp = {}
    for i, val in enumerate(a):
        need = target - val
        if need in mp:
            print(mp[need] + 1, i + 1)
            return
        if val not in mp:
            mp[val] = i

    print("-1 -1")

if __name__ == "__main__":
    main()
