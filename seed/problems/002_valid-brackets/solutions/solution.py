import sys


def main():
    s = sys.stdin.readline().strip()

    stack = []
    matching = {')': '(', ']': '[', '}': '{'}

    for c in s:
        if c in '([{':
            stack.append(c)
        elif c in ')]}':
            if not stack or stack[-1] != matching[c]:
                print("No")
                return
            stack.pop()

    print("Yes" if not stack else "No")


if __name__ == "__main__":
    main()
