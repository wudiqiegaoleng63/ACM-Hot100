#include <iostream>
#include <stack>
#include <string>

using namespace std;

int main() {
    string s;
    getline(cin, s);

    stack<char> stk;
    for (char c : s) {
        if (c == '(' || c == '[' || c == '{') {
            stk.push(c);
        } else if (c == ')' || c == ']' || c == '}') {
            if (stk.empty()) {
                cout << "No" << endl;
                return 0;
            }
            char top = stk.top();
            stk.pop();
            if ((c == ')' && top != '(') ||
                (c == ']' && top != '[') ||
                (c == '}' && top != '{')) {
                cout << "No" << endl;
                return 0;
            }
        }
    }

    if (stk.empty()) {
        cout << "Yes" << endl;
    } else {
        cout << "No" << endl;
    }

    return 0;
}
