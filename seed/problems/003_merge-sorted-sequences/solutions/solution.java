import java.io.*;
import java.util.*;

public class Main {
    public static void main(String[] args) throws IOException {
        BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
        StringTokenizer st = new StringTokenizer(br.readLine());
        int n = Integer.parseInt(st.nextToken());
        int m = Integer.parseInt(st.nextToken());

        int[] a = new int[n];
        st = new StringTokenizer(br.readLine());
        for (int i = 0; i < n; i++) {
            a[i] = Integer.parseInt(st.nextToken());
        }

        int[] b = new int[m];
        st = new StringTokenizer(br.readLine());
        for (int i = 0; i < m; i++) {
            b[i] = Integer.parseInt(st.nextToken());
        }

        StringBuilder sb = new StringBuilder();
        int i = 0, j = 0;
        while (i < n && j < m) {
            if (a[i] <= b[j]) {
                if (sb.length() > 0) sb.append(' ');
                sb.append(a[i++]);
            } else {
                if (sb.length() > 0) sb.append(' ');
                sb.append(b[j++]);
            }
        }
        while (i < n) {
            if (sb.length() > 0) sb.append(' ');
            sb.append(a[i++]);
        }
        while (j < m) {
            if (sb.length() > 0) sb.append(' ');
            sb.append(b[j++]);
        }
        sb.append('\n');

        System.out.print(sb);
    }
}
