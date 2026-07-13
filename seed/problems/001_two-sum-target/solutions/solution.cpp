#include <iostream>
#include <unordered_map>
#include <vector>
using namespace std;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int n;
    long long target;
    cin >> n >> target;

    vector<long long> a(n);
    for (int i = 0; i < n; i++) {
        cin >> a[i];
    }

    unordered_map<long long, int> mp;
    for (int i = 0; i < n; i++) {
        long long need = target - a[i];
        auto it = mp.find(need);
        if (it != mp.end()) {
            cout << it->second + 1 << " " << i + 1 << "\n";
            return 0;
        }
        if (mp.find(a[i]) == mp.end()) {
            mp[a[i]] = i;
        }
    }

    cout << "-1 -1\n";
    return 0;
}
