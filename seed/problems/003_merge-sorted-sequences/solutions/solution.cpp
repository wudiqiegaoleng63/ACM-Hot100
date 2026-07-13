#include <iostream>
#include <vector>
using namespace std;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int n, m;
    cin >> n >> m;

    vector<int> a(n), b(m);
    for (int i = 0; i < n; i++) cin >> a[i];
    for (int i = 0; i < m; i++) cin >> b[i];

    vector<int> result;
    result.reserve(n + m);

    int i = 0, j = 0;
    while (i < n && j < m) {
        if (a[i] <= b[j]) {
            result.push_back(a[i++]);
        } else {
            result.push_back(b[j++]);
        }
    }
    while (i < n) result.push_back(a[i++]);
    while (j < m) result.push_back(b[j++]);

    for (int k = 0; k < (int)result.size(); k++) {
        if (k > 0) cout << ' ';
        cout << result[k];
    }
    cout << '\n';

    return 0;
}
