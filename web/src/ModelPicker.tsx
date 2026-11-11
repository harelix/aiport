import React, {useEffect, useId, useState} from 'react';

const OTHER = '\u0000other';

/** The models a Connection lists (GET /api/v1/targets/{id}/models), loaded when the Connection is chosen. */
export function useConnectionModels(connection: string): {models: string[]; state: 'loading' | 'ok' | 'unlisted'} {
  const [result, setResult] = useState<{models: string[]; state: 'loading' | 'ok' | 'unlisted'}>({models: [], state: 'loading'});
  useEffect(() => {
    if (!connection) { setResult({models: [], state: 'unlisted'}); return; }
    const controller = new AbortController();
    setResult({models: [], state: 'loading'});
    fetch('/api/v1/targets/' + encodeURIComponent(connection) + '/models', {credentials: 'same-origin', headers: {'X-Aiport-CSRF': '1'}, signal: controller.signal})
      .then(r => r.ok ? r.json() : {}).then((d: {models?: string[]}) => { if (!controller.signal.aborted) setResult(d.models?.length ? {models: [...d.models].sort(), state: 'ok'} : {models: [], state: 'unlisted'}); })
      .catch(() => { if (!controller.signal.aborted) setResult({models: [], state: 'unlisted'}); });
    return () => controller.abort();
  }, [connection]);
  return result;
}

/* A model choice: the Connection's listed models, or Other for an id it does
   not list. empty, when set, is the first option's label (an empty value). */
export function ModelPicker({id, label, value, onChange, models, state, empty, required}: {id: string; label: string; value: string; onChange: (v: string) => void; models: string[]; state: string; empty?: string; required?: boolean}) {
  const listed = models.includes(value);
  const [other, setOther] = useState(false);
  const typing = state !== 'ok' || other || (!!value && !listed);
  return <label className="model-picker" htmlFor={id}>{label}
    {state === 'ok' && <select id={typing ? undefined : id} value={typing ? OTHER : value} required={required && !typing} onChange={e => { if (e.target.value === OTHER) { setOther(true); onChange(listed ? '' : value); } else { setOther(false); onChange(e.target.value); } }}>
      {empty !== undefined ? <option value="">{empty}</option> : !value && <option value="" disabled>Choose a model</option>}
      {models.map(m => <option key={m} value={m}>{m}</option>)}
      <option value={OTHER}>Other model id…</option>
    </select>}
    {typing && <input id={id} value={value} onChange={e => onChange(e.target.value)} required={required} maxLength={256} spellCheck={false} autoComplete="off"
      placeholder={state === 'loading' ? 'Loading the models…' : empty !== undefined ? empty : 'The model id the provider expects'}/>}
  </label>;
}
/** Load the selected Connection's models wherever a form chooses a provider and model. */
export function ConnectionModelPicker({connection, label = 'Model', value, onChange, required = true}: {connection: string; label?: string; value: string; onChange: (v: string) => void; required?: boolean}) {
  const id = useId();
  const listed = useConnectionModels(connection);
  return <div>
    {connection ? <ModelPicker key={connection} id={id} label={label} value={value} onChange={onChange} models={listed.models} state={listed.state} required={required}/> : <label htmlFor={id}>{label}<input id={id} value={value} disabled/></label>}
    {connection && listed.state === 'unlisted' && <small role="status">The model list is unavailable or empty. Enter the model id this Connection serves.</small>}
  </div>;
}
