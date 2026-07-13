import sys
from collections import deque

def main():
    data = sys.stdin.read().split()
    n = int(data[0])
    a = [int(x) for x in data[1:n+1]]

    # BFS approach: process level by level using index-based traversal
    # Node at index i has left child at 2i+1 and right child at 2i+2
    q = deque()
    if n > 0 and a[0] != -1:
        q.append(0)

    output_lines = []
    while q:
        level_size = len(q)
        level_values = []

        for _ in range(level_size):
            idx = q.popleft()
            level_values.append(a[idx])

            left = 2 * idx + 1
            right = 2 * idx + 2
            if left < n and a[left] != -1:
                q.append(left)
            if right < n and a[right] != -1:
                q.append(right)

        output_lines.append(' '.join(str(v) for v in level_values))

    print('\n'.join(output_lines))

if __name__ == '__main__':
    main()
