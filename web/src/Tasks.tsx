import React, {FormEvent, useEffect, useState} from 'react';
import {Check, Copy, ListTodo, X} from 'lucide-react';

type Status = 'pending' | 'in_progress' | 'blocked' | 'completed' | 'failed' | 'cancelled';
type Entry = {at_ms: number; actor: string; action: string; from?: string; to?: string; note?: string};
type Task = {uuid: string; name: string; title: string; instructions: string; acceptance: string; status: Status; priority: number; created_at_ms: number; updated_at_ms: number; history: Entry[]};

const statusLabel: Record<Status, string> = {pending: 'Pending', in_progress: 'In progress', blocked: 'Blocked', completed: 'Completed', failed: 'Failed', cancelled: 'Cancelled'};
const statusTone: Record<Status, string> = {pending: '', in_progress: 'info', blocked: 'warn', completed: 'good', failed: 'bad', cancelled: ''};
const priorityLabel = ['None', 'Low', 'Medium', 'High'];
const when = (ms: number) => new Date(ms).toLocaleString(undefined, {dateStyle: 'medium', timeStyle: 'short'});

async function api<T>(path: string, method = 'GET', body?: object): Promise<T> {
  const response = await fetch('/api/v1/tasks' + path, {
    method, credentials: 'same-origin',
    headers: method === 'GET' ? {} : {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'},
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error?.message || 'Task request failed (' + response.status + ').');
  return data;
}

function StatusBadge({status}: {status: Status}) {
  return <span className={'badge ' + statusTone[status]}>{statusLabel[status]}</span>;
}

function CopyName({name}: {name: string}) {
  const [copied, setCopied] = useState(false);
  return <button type="button" className="task-name" title="Copy task name" onClick={e => {
    e.stopPropagation();
    void navigator.clipboard?.writeText(name).then(() => { setCopied(true); setTimeout(() => setCopied(false), 1200); });
  }}>{copied ? <><Check size={12} aria-hidden/> Copied</> : <><Copy size={12} aria-hidden/> {name}</>}</button>;
}

function NewTask({onCreated}: {onCreated: (t: Task) => void}) {
  const [title, setTitle] = useState(''), [instructions, setInstructions] = useState(''), [acceptance, setAcceptance] = useState(''), [priority, setPriority] = useState(0);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [open, setOpen] = useState(false);
  async function submit(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError('');
    try {
      const t = await api<Task>('', 'POST', {title, instructions, acceptance, priority});
      setTitle(''); setInstructions(''); setAcceptance(''); setPriority(0); setOpen(false);
      onCreated(t);
    } catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  return <section className="panel task-new">
    <form onSubmit={submit}>
      <div className="task-quick">
        <input aria-label="New task title" placeholder="Capture a task… e.g. Add keyboard shortcuts" value={title} onChange={e => setTitle(e.target.value)} onFocus={() => setOpen(true)} maxLength={200} required/>
        <button className="primary" disabled={busy || !title.trim()}>{busy ? 'Adding…' : 'Add task'}</button>
      </div>
      {open && <div className="task-more">
        <label>Instructions<textarea rows={4} value={instructions} onChange={e => setInstructions(e.target.value)} placeholder="What needs to change, and any constraints" maxLength={20000}/></label>
        <label>Acceptance criteria<textarea rows={3} value={acceptance} onChange={e => setAcceptance(e.target.value)} placeholder="How completion is verified" maxLength={10000}/></label>
        <label className="task-priority">Priority<select value={priority} onChange={e => setPriority(Number(e.target.value))}>{priorityLabel.map((p, i) => <option key={p} value={i}>{p}</option>)}</select></label>
      </div>}
      {error && <p className="auth-error" role="alert">{error}</p>}
    </form>
  </section>;
}

function TaskDetail({task, onChange, onClose}: {task: Task; onChange: (t: Task) => void; onClose: () => void}) {
  const [title, setTitle] = useState(task.title), [instructions, setInstructions] = useState(task.instructions), [acceptance, setAcceptance] = useState(task.acceptance), [priority, setPriority] = useState(task.priority);
  const [status, setStatus] = useState<Status>(task.status), [note, setNote] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  useEffect(() => { setTitle(task.title); setInstructions(task.instructions); setAcceptance(task.acceptance); setPriority(task.priority); setStatus(task.status); setNote(''); }, [task.uuid, task.updated_at_ms]);
  async function save(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError(''); setMessage('');
    try {
      const t = await api<Task>('/' + encodeURIComponent(task.name), 'PATCH', {title, instructions, acceptance, priority, status, ...(note.trim() ? {note} : {})});
      setMessage('Saved.'); onChange(t);
    } catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  return <section className="panel task-detail" aria-label={'Task ' + task.name}>
    <div className="panel-heading"><div><CopyName name={task.name}/><h2>{task.title}</h2></div><button className="text-button" onClick={onClose}>Close <X size={14} aria-hidden/></button></div>
    <form onSubmit={save} className="task-form">
      <label>Title<input value={title} onChange={e => setTitle(e.target.value)} maxLength={200} required/></label>
      <div className="task-row">
        <label>Status<select value={status} onChange={e => setStatus(e.target.value as Status)}>{(Object.keys(statusLabel) as Status[]).map(s => <option key={s} value={s}>{statusLabel[s]}</option>)}</select></label>
        <label>Priority<select value={priority} onChange={e => setPriority(Number(e.target.value))}>{priorityLabel.map((p, i) => <option key={p} value={i}>{p}</option>)}</select></label>
      </div>
      <label>Instructions<textarea rows={6} value={instructions} onChange={e => setInstructions(e.target.value)} maxLength={20000}/></label>
      <label>Acceptance criteria<textarea rows={4} value={acceptance} onChange={e => setAcceptance(e.target.value)} maxLength={10000}/></label>
      <label>Note for history (optional)<input value={note} onChange={e => setNote(e.target.value)} maxLength={2000} placeholder="e.g. blocked on API design"/></label>
      {error && <p className="auth-error" role="alert">{error}</p>}
      {message && <p className="auth-success" role="status">{message}</p>}
      <button className="primary" disabled={busy}>{busy ? 'Saving…' : 'Save changes'}</button>
    </form>
    <div className="task-history">
      <h3>History</h3>
      <ol>{[...task.history].reverse().map((h, i) => <li key={i}>
        <span className="task-when">{when(h.at_ms)}</span>
        <span><strong>{h.actor}</strong> {h.action === 'created' ? 'created the task' : h.action === 'status' ? <>changed status {statusLabel[h.from as Status] || h.from} → {statusLabel[h.to as Status] || h.to}</> : h.action === 'edited' ? 'edited ' + (h.note || '') : 'added a note'}
          {h.action !== 'edited' && h.note && <em> · {h.note}</em>}</span>
      </li>)}</ol>
    </div>
  </section>;
}

export default function Tasks({refresh}: {refresh: number}) {
  const [tasks, setTasks] = useState<Task[]>([]), [filter, setFilter] = useState<'open' | 'all' | Status>('open'), [selected, setSelected] = useState<string>(''), [error, setError] = useState(''), [loading, setLoading] = useState(true);
  useEffect(() => {
    setLoading(true);
    api<{tasks: Task[]}>('').then(d => { setTasks(d.tasks); setError(''); }).catch(e => setError((e as Error).message)).finally(() => setLoading(false));
  }, [refresh]);
  const upsert = (t: Task) => setTasks(list => [t, ...list.filter(x => x.uuid !== t.uuid)].sort((a, b) => b.created_at_ms - a.created_at_ms));
  const open = (s: Status) => s === 'pending' || s === 'in_progress' || s === 'blocked';
  const shown = tasks.filter(t => filter === 'all' || (filter === 'open' ? open(t.status) : t.status === filter));
  const counts = (Object.keys(statusLabel) as Status[]).map(s => [s, tasks.filter(t => t.status === s).length] as const);
  const current = tasks.find(t => t.uuid === selected);
  return <>
    {error && <div className="notice error" role="alert">{error}</div>}
    <NewTask onCreated={t => { upsert(t); setSelected(t.uuid); }}/>
    <div className="task-filters" role="tablist" aria-label="Filter tasks">
      {([['open', 'Open', tasks.filter(t => open(t.status)).length], ['all', 'All', tasks.length], ...counts.map(([s, n]) => [s, statusLabel[s], n])] as [string, string, number][]).map(([key, label, n]) =>
        <button key={key} role="tab" aria-selected={filter === key} className={filter === key ? 'active' : ''} onClick={() => setFilter(key as typeof filter)}>{label}<span>{n}</span></button>)}
    </div>
    <div className={current ? 'task-layout with-detail' : 'task-layout'}>
      <section className="panel">
        <div className="panel-heading"><h2>Task queue</h2><span>{shown.length} shown · you own intent and priority</span></div>
        {loading ? <div className="quiet-empty">Loading tasks…</div> : shown.length ? <div className="table-scroll"><table>
          <thead><tr><th>Task</th><th>Status</th><th>Priority</th><th className="numeric">Updated</th></tr></thead>
          <tbody>{shown.map(t => <tr key={t.uuid} className={'task-row-item' + (t.uuid === selected ? ' selected' : '')} onClick={() => setSelected(t.uuid)} tabIndex={0} onKeyDown={e => { if (e.key === 'Enter') setSelected(t.uuid); }}>
            <td><span className="task-title">{t.title}</span><CopyName name={t.name}/></td>
            <td><StatusBadge status={t.status}/></td>
            <td>{priorityLabel[t.priority]}</td>
            <td className="numeric"><small>{when(t.updated_at_ms)}</small></td>
          </tr>)}</tbody>
        </table></div> : <div className="empty"><span className="empty-icon"><ListTodo size={20} strokeWidth={1.75} aria-hidden/></span><h3>{tasks.length ? 'Nothing here' : 'No tasks yet'}</h3><p>{tasks.length ? 'No tasks match this filter.' : 'Capture work as it comes up. Each task gets a memorable name like task-amber-fox that you can hand to a coding agent.'}</p></div>}
      </section>
      {current && <TaskDetail task={current} onChange={upsert} onClose={() => setSelected('')}/>}
    </div>
  </>;
}
