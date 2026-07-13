import sys

def main():
    data = sys.stdin.read().split()
    n = int(data[0])
    a = list(map(int, data[1:n+1]))

    max_sum = a[0]
    current_sum = a[0]
    for i in range(1, n):
        current_sum = max(a[i], current_sum + a[i])
        max_sum = max(max_sum, current_sum)

    print(max_sum)

if __name__ == "__main__":
    main()
