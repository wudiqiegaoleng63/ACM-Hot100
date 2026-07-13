import java.io.*;
import java.util.*;

public class Main {
    public static void main(String[] args) throws IOException {
        BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
        StringTokenizer st = new StringTokenizer(br.readLine());
        int n = Integer.parseInt(st.nextToken());
        long target = Long.parseLong(st.nextToken());

        long[] a = new long[n];
        st = new StringTokenizer(br.readLine());
        for (int i = 0; i < n; i++) {
            a[i] = Long.parseLong(st.nextToken());
        }

        Map<Long, Integer> map = new HashMap<>();
        for (int i = 0; i < n; i++) {
            long need = target - a[i];
            if (map.containsKey(need)) {
                System.out.println((map.get(need) + 1) + " " + (i + 1));
                return;
            }
            if (!map.containsKey(a[i])) {
                map.put(a[i], i);
            }
        }

        System.out.println("-1 -1");
    }
}
