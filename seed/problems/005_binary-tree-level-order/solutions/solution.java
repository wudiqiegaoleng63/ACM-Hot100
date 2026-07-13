import java.io.*;
import java.util.*;

public class Main {
    public static void main(String[] args) throws IOException {
        BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
        int n = Integer.parseInt(br.readLine().trim());
        StringTokenizer st = new StringTokenizer(br.readLine());
        int[] a = new int[n];
        for (int i = 0; i < n; i++) {
            a[i] = Integer.parseInt(st.nextToken());
        }

        // BFS approach: process level by level using index-based traversal
        // Node at index i has left child at 2i+1 and right child at 2i+2
        Deque<Integer> queue = new ArrayDeque<>();
        if (n > 0 && a[0] != -1) {
            queue.addLast(0);
        }

        StringBuilder sb = new StringBuilder();
        boolean firstLine = true;
        while (!queue.isEmpty()) {
            int levelSize = queue.size();
            if (!firstLine) sb.append('\n');
            firstLine = false;

            for (int i = 0; i < levelSize; i++) {
                int idx = queue.pollFirst();
                if (i > 0) sb.append(' ');
                sb.append(a[idx]);

                int left = 2 * idx + 1;
                int right = 2 * idx + 2;
                if (left < n && a[left] != -1) {
                    queue.addLast(left);
                }
                if (right < n && a[right] != -1) {
                    queue.addLast(right);
                }
            }
        }

        System.out.println(sb);
    }
}
