import React, {useEffect, useState} from 'react';
import {Activity, Bot, Boxes, History, LayoutDashboard, ListTodo, LogOut, Monitor, Moon, PanelLeftClose, PanelLeftOpen, Plug, Route, Settings, ShieldOff, Shrink, Sun, ThumbsUp, type LucideIcon} from 'lucide-react';

export type Appearance = 'light' | 'dark' | 'system';
const APPEARANCE_KEY = 'aiport.appearance';
const COLLAPSED_KEY = 'aiport.sidebar.collapsed';

function stored<T extends string>(key: string, allowed: readonly T[], fallback: T): T {
  try {
    const v = localStorage.getItem(key);
    return allowed.includes(v as T) ? (v as T) : fallback;
  } catch {
    return fallback;
  }
}
function save(key: string, value: string) {
  try { localStorage.setItem(key, value); } catch { /* storage unavailable: preference lasts this session */ }
}

/** Appearance preference (Light / Dark / System), persisted locally; System follows the OS live. */
export function useAppearance(): [Appearance, (a: Appearance) => void] {
  const [appearance, setAppearance] = useState<Appearance>(() => stored(APPEARANCE_KEY, ['light', 'dark', 'system'] as const, 'system'));
  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: light)');
    const apply = () => {
      document.documentElement.dataset.theme = appearance === 'system' ? (media.matches ? 'light' : 'dark') : appearance;
    };
    apply();
    save(APPEARANCE_KEY, appearance);
    if (appearance !== 'system') return;
    media.addEventListener('change', apply);
    return () => media.removeEventListener('change', apply);
  }, [appearance]);
  return [appearance, setAppearance];
}

const appearanceOptions: [Appearance, string, LucideIcon][] = [['light', 'Light', Sun], ['dark', 'Dark', Moon], ['system', 'System', Monitor]];
/** Light / Dark / System switch, icon-only, for the page header. */
export function ThemeSwitch({appearance, onAppearance}: {appearance: Appearance; onAppearance: (a: Appearance) => void}) {
  return <div className="theme-switch" role="radiogroup" aria-label="Appearance">
    {appearanceOptions.map(([value, label, Icon]) =>
      <button key={value} type="button" role="radio" aria-checked={appearance === value} aria-label={label} title={label + ' appearance'} className={appearance === value ? 'on' : ''} onClick={() => onAppearance(value)}>
        <Icon size={15} strokeWidth={1.75} aria-hidden/>
      </button>)}
  </div>;
}

// Ordered by how often the Owner needs them: watch and decide first, then set up.
// Changes sits with the daily pages; it stays quiet, with a badge only when something needs a decision.
const trafficPages: [string, LucideIcon][] = [['Overview', LayoutDashboard], ['Activity', Activity], ['Changes', History]];
const gatewayPages: [string, LucideIcon][] = [['Integrate', Plug], ['Providers', Boxes], ['Routing', Route], ['Agents', Bot], ['Compression', Shrink], ['Feedback', ThumbsUp]];
const utilityPages: [string, LucideIcon][] = [['Tasks', ListTodo], ['Settings', Settings]];

type Props = {
  logo: string; page: string; onPage: (p: string) => void; listen?: string;
  authDisabled: boolean; onSignOut: () => void;
  /** How many items in Changes wait for the Owner's decision (badge shown only when > 0). */
  pending?: number;
};

export default function Sidebar({logo, page, onPage, listen, authDisabled, onSignOut, pending = 0}: Props) {
  const [collapsed, setCollapsed] = useState(() => stored(COLLAPSED_KEY, ['1', '0'] as const, '0') === '1');
  useEffect(() => {
    save(COLLAPSED_KEY, collapsed ? '1' : '0');
    document.documentElement.dataset.sidebar = collapsed ? 'collapsed' : 'expanded';
  }, [collapsed]);
  const navButton = ([name, Icon]: [string, LucideIcon]) => {
    const count = name === 'Changes' ? pending : 0;
    const label = count ? name + ', ' + count + ' waiting for a decision' : name;
    return <button key={name} className={page === name ? 'active' : ''} aria-current={page === name ? 'page' : undefined} title={collapsed ? label : undefined} aria-label={collapsed || count ? label : undefined} onClick={() => onPage(name)}>
      <Icon size={18} strokeWidth={1.75} aria-hidden/><span className="label">{name}</span>
      {!!count && <span className="nav-badge" aria-hidden>{count}</span>}
    </button>;
  };
  return <aside className={'sidebar' + (collapsed ? ' collapsed' : '')}>
    <div className="sidebar-top">
      <div className="brand"><img className="service-logo" src={logo} alt="" width="88" height="88"/><span className="label">aiport</span></div>
      <button className="icon-button collapse-toggle" onClick={() => setCollapsed(c => !c)} aria-expanded={!collapsed} aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'} title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}>
        {collapsed ? <PanelLeftOpen size={18} strokeWidth={1.75} aria-hidden/> : <PanelLeftClose size={18} strokeWidth={1.75} aria-hidden/>}
      </button>
    </div>
    <nav aria-label="Main navigation">
      <span className="nav-group label">Monitor</span>{trafficPages.map(navButton)}
      <span className="nav-group label">Configure</span>{gatewayPages.map(navButton)}
    </nav>
    <div className="sidebar-bottom">
      <nav aria-label="Workspace">{utilityPages.map(navButton)}</nav>
      <div className="sidebar-foot">
        <div className="local-label" title={listen}><span className="dot"/><span className="label">{listen || 'Connecting…'}</span></div>
        {authDisabled
          ? <button className="icon-text" onClick={() => onPage('Settings')} title="Authentication is off" aria-label="Authentication is off. Open Settings"><ShieldOff size={16} strokeWidth={1.75} aria-hidden/><span className="label">Authentication off</span></button>
          : <button className="icon-text" onClick={onSignOut} title="Sign out" aria-label="Sign out"><LogOut size={16} strokeWidth={1.75} aria-hidden/><span className="label">Sign out</span></button>}
      </div>
    </div>
  </aside>;
}
