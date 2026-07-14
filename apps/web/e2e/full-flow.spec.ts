import { expect, test } from '@playwright/test';

const mailpitURL = process.env.E2E_MAILPIT_URL ?? 'http://127.0.0.1:8025';
const runID = process.env.E2E_RUN_ID ?? Date.now().toString(36);
const email = `e2e-${runID}@example.local`;
const username = `e2e_${runID}`.slice(0, 20);
const password = 'E2ePassword123!';

test.setTimeout(120_000);

test('registration through accepted submission detail', async ({ page, request }) => {
  await request.delete(`${mailpitURL}/api/v1/messages`);

  await page.goto('/register');
  await page.getByPlaceholder('your@email.com').fill(email);
  await page.getByPlaceholder('username').fill(username);
  await page.getByPlaceholder('设置密码').fill(password);
  await page.getByPlaceholder('再次输入密码').fill(password);
  await page.getByRole('button', { name: '注册', exact: true }).click();
  await expect(page.getByRole('heading', { name: '注册成功' })).toBeVisible();

  let verificationLink = '';
  await expect.poll(async () => {
    const response = await request.get(`${mailpitURL}/api/v1/messages`);
    if (!response.ok()) return '';
    const payload = await response.json() as MailpitList;
    const message = payload.messages.find((item) => item.To.some((recipient) => recipient.Address === email));
    if (!message) return '';
    const detailResponse = await request.get(`${mailpitURL}/api/v1/message/${message.ID}`);
    if (!detailResponse.ok()) return '';
    const detail = await detailResponse.json() as MailpitMessage;
    const match = detail.HTML.match(/href="(https?:\/\/[^\"]+\/verify-email\?token=[^\"]+)"/);
    verificationLink = match?.[1] ?? '';
    return verificationLink;
  }, { timeout: 10_000, intervals: [200, 500, 1_000] }).not.toBe('');

  await page.goto(verificationLink);
  await expect(page.getByRole('heading', { name: '邮箱验证成功' })).toBeVisible();
  await page.getByRole('link', { name: '前往登录' }).click();

  await page.getByPlaceholder('your@email.com').fill(email);
  await page.getByPlaceholder('输入密码').fill(password);
  await page.getByRole('button', { name: '登录', exact: true }).click();
  await expect(page).toHaveURL(/\/problems$/);
  await expect(page.getByRole('heading', { name: 'ACM Hot 100' })).toBeVisible();

  await page.getByRole('link', { name: '两数目标和' }).click();
  await expect(page).toHaveURL(/\/problems\/two-sum-target$/);
  await expect(page.getByRole('heading', { name: '两数目标和' }).first()).toBeVisible();
  await expect(page.getByText('Mock Judge')).toBeVisible();

  const draftSource = '#include <iostream>\nint main(){ std::cout << 3 << "\\n"; }';
  const languageKey = await page.getByLabel('编程语言').inputValue();
  const userID = await page.evaluate(async () => {
    const response = await fetch('/api/v1/auth/me', { credentials: 'include' });
    const payload = await response.json() as { user: { id: string } };
    return payload.user.id;
  });
  await page.evaluate(({ key, source }) => {
    localStorage.setItem(key, JSON.stringify({ source_code: source, updated_at: new Date().toISOString() }));
  }, { key: `draft:${userID}:two-sum-target:${languageKey}`, source: draftSource });
  await page.reload();
  await expect.poll(async () => page.evaluate(() => {
    const keys = Object.keys(localStorage).filter((key) => key.startsWith('draft:'));
    return keys.some((key) => JSON.parse(localStorage.getItem(key) ?? '{}').source_code?.includes('std::cout'));
  })).toBe(true);

  await page.getByRole('button', { name: '运行样例' }).click();
  await expect(page.getByText('通过', { exact: true })).toBeVisible({ timeout: 10_000 });

  await page.getByRole('button', { name: '正式提交' }).click();
  await expect(page.getByText('答案正确')).toBeVisible({ timeout: 15_000 });

  await page.goto('/profile');
  await expect(page.getByRole('heading', { name: '个人进度' })).toBeVisible();
  await expect(page.getByText('1 / 5')).toBeVisible();

  await page.goto('/submissions');
  await expect(page.getByRole('heading', { name: '我的提交' })).toBeVisible();
  await page.getByRole('link', { name: '两数目标和' }).first().click();
  await expect(page).toHaveURL(/\/submissions\/[^/]+$/);
  await expect(page.getByRole('heading', { name: '提交代码' })).toBeVisible();
  await expect(page.getByText('答案正确')).toBeVisible();
  await expect(page.getByRole('link', { name: '回到题目继续修改' })).toBeVisible();
});

interface MailpitList {
  messages: Array<{
    ID: string;
    To: Array<{ Address: string }>;
  }>;
}

interface MailpitMessage {
  HTML: string;
}
