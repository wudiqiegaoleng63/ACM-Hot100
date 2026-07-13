import sys

def main():
    data = sys.stdin.read().split()
    idx = 0
    n = int(data[idx]); idx += 1
    m = int(data[idx]); idx += 1

    a = [int(data[idx + i]) for i in range(n)]; idx += n
    b = [int(data[idx + i]) for i in range(m)]; idx += m

    result = []
    i, j = 0, 0
    while i < n and j < m:
        if a[i] <= b[j]:
            result.append(a[i])
            i += 1
        else:
            result.append(b[j])
            j += 1
    while i < n:
        result.append(a[i])
        i += 1
    while j < m:
        result.append(b[j])
        j += 1

    print(' '.join(map(str, result)))

if __name__ == '__main__':
    main()
