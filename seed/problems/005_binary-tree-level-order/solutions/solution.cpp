#include <bits/stdc++.h>
using namespace std;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int n;
    cin >> n;
    vector<int> a(n);
    for (int i = 0; i < n; i++) {
        cin >> a[i];
    }

    // BFS approach: process level by level using index-based traversal
    // Node at index i has left child at 2i+1 and right child at 2i+2
    queue<int> q;
    if (n > 0 && a[0] != -1) {
        q.push(0);
    }

    bool first_line = true;
    while (!q.empty()) {
        int level_size = q.size();
        vector<int> level_values;

        for (int i = 0; i < level_size; i++) {
            int idx = q.front();
            q.pop();
            level_values.push_back(a[idx]);

            int left = 2 * idx + 1;
            int right = 2 * idx + 2;
            if (left < n && a[left] != -1) {
                q.push(left);
            }
            if (right < n && a[right] != -1) {
                q.push(right);
            }
        }

        if (!first_line) cout << '\n';
        for (int i = 0; i < (int)level_values.size(); i++) {
            if (i > 0) cout << ' ';
            cout << level_values[i];
        }
        first_line = false;
    }

    cout << '\n';
    return 0;
}
