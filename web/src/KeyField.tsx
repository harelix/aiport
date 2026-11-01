import React, {FormEvent, useState} from 'react';

export type KeySource = 'saved' | 'environment' | 'missing' | 'not_needed';
type KeyResult = {secret_status: Record<string, boolean>; secret_source: Record<string, KeySource>};

async function keyRequest(target: string, method: 'PUT' | 'DELETE', apiKey?: string): Promise<KeyResult> {
  const response = await fetch('/api/v1/targets/' + encodeURIComponent(target) + '/api-key', {
    method,
    credentials: 'same-origin',
    headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'},
    body: method === 'PUT' ? JSON.stringify({api_key: apiKey}) : undefined,
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error?.message || 'Could not update the API key.');
  return data;
}

export const keyBadge: Record<KeySource, [string, string]> = {
  saved: ['Key saved', 'good'],
  environment: ['Key from environment', 'good'],
  missing: ['Key missing', 'warn'],
  not_needed: ['No key needed', 'good'],
};

/**
 * Write-only provider API key field. The key is sent once to the local server,
 * stored in the private credential file, and never shown again.
 */
export default function KeyField({target, envName, source, onChange}: {target: string; envName: string; source: KeySource; onChange: (r: KeyResult) => void}) {
  const [value, setValue] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  async function run(method: 'PUT' | 'DELETE') {
    setBusy(true); setError(''); setMessage('');
    try {
      const result = await keyRequest(target, method, value.trim());
      setValue('');
      setMessage(method === 'PUT' ? 'Saved. Applies to new requests immediately.' : result.secret_source[target] === 'environment' ? 'Removed. Using ' + envName + ' from the server environment.' : 'Removed.');
      onChange(result);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  const submit = (e: FormEvent) => { e.preventDefault(); if (value.trim()) void run('PUT'); };
  const id = 'key-' + target;
  return <form className="key-field" onSubmit={submit}>
    <label htmlFor={id}>{source === 'missing' ? 'API key' : 'Replace API key'}</label>
    <div className="key-row">
      <input id={id} type="password" autoComplete="off" spellCheck={false} value={value} onChange={e => setValue(e.target.value)}
        placeholder={source === 'saved' ? '•••••••• saved — paste to replace' : source === 'environment' ? 'Using ' + envName + ' — paste to override' : 'Paste your ' + envName}
        maxLength={4000}/>
      <button className="secondary" disabled={busy || !value.trim()}>{busy ? 'Saving…' : 'Save key'}</button>
    </div>
    {error && <p className="auth-error" role="alert">{error}</p>}
    {message && <p className="auth-success" role="status">{message}</p>}
    {source === 'saved' && <button type="button" className="text-button key-remove" disabled={busy} onClick={() => void run('DELETE')}>Remove saved key</button>}
  </form>;
}
