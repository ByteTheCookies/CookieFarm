import { Resend } from 'resend';

const RESEND_API_KEY = process.env.RESEND_API_KEY || ''; /^re_[\w]+$/.test(RESEND_API_KEY) || (() => { throw new Error('Invalid RESEND_API_KEY'); })();
const resend = new Resend(RESEND_API_KEY);

const email_list = ['akiidjk@bytethecookies.org', 'suga@bytethecookes.org','giovanni@bytethecookes.org'];

export async function sendEmail({ url, opinion, message }: { url: string; opinion: string; message: string }) {
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

  const { error } = await resend.emails.send({
    from: 'CookieFarm Docs <cookiefarm@bytethecookies.org>',
    to: email_list,
    subject: `Feedback Received - ${opinion}`,
    html,
  });

  if (error) {
    return console.error({ error });
  }
}
