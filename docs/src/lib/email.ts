import { Resend } from 'resend';

let resend: Resend | undefined;

function getResend(): Resend {
  if (!resend) {
    const key = process.env.RESEND_API_KEY ?? '';
    if (!/^re_\w+$/.test(key)) throw new Error('Invalid RESEND_API_KEY');
    resend = new Resend(key);
  }
  return resend;
}

// Feedback fields come from anonymous visitors: escape before putting them in HTML.
function escapeHtml(s: string): string {
  return s
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

const email_list = ['akiidjk@bytethecookies.org', 'suga@bytethecookies.org', 'giovanni@bytethecookies.org'];

export async function sendEmail(feedback: { url: string; opinion: string; message: string }) {
  const url = escapeHtml(feedback.url);
  const opinion = escapeHtml(feedback.opinion);
  const message = escapeHtml(feedback.message);

  const html = `
    <html>
      <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Feedback Received</title>
      </head>
      <body style="margin:0;padding:24px;background:#f6f8fb;font-family:Arial,Helvetica,sans-serif;color:#1f2937;">
        <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:640px;margin:0 auto;background:#ffffff;border:1px solid #e5e7eb;border-radius:12px;overflow:hidden;">
          <tr>
            <td style="padding:20px 24px;background:#111827;color:#ffffff;">
              <h1 style="margin:0;font-size:20px;line-height:1.4;">New Feedback Received</h1>
            </td>
          </tr>
          <tr>
            <td style="padding:24px;">
              <p style="margin:0 0 16px 0;font-size:14px;color:#4b5563;">A new piece of feedback was submitted on CookieFarm Docs.</p>
              <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="border-collapse:collapse;font-size:14px;">
                <tr>
                  <td style="padding:10px 12px;border:1px solid #e5e7eb;background:#f9fafb;width:120px;"><strong>URL</strong></td>
                  <td style="padding:10px 12px;border:1px solid #e5e7eb;word-break:break-word;">${url}</td>
                </tr>
                <tr>
                  <td style="padding:10px 12px;border:1px solid #e5e7eb;background:#f9fafb;"><strong>Opinion</strong></td>
                  <td style="padding:10px 12px;border:1px solid #e5e7eb;text-transform:capitalize;">${opinion}</td>
                </tr>
                <tr>
                  <td style="padding:10px 12px;border:1px solid #e5e7eb;background:#f9fafb;vertical-align:top;"><strong>Message</strong></td>
                  <td style="padding:10px 12px;border:1px solid #e5e7eb;white-space:pre-wrap;">${message}</td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
      </body>
    </html>
  `;

  const { error } = await getResend().emails.send({
    from: 'CookieFarm Docs <cookiefarm@bytethecookies.org>',
    to: email_list,
    subject: `Feedback Received - ${opinion}`,
    html,
  });

  if (error) {
    return console.error({ error });
  }
}
