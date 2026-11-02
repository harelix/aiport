import React, {createContext, useContext} from 'react';
import {Info} from 'lucide-react';

/** Opens the built-in manual at a section (its heading text, exactly as in docs/manual.md). */
export const HelpContext = createContext<(section: string) => void>(() => {});

/** The manual section that explains each page, for the info button next to the page title. */
export const pageHelp: Record<string, string> = {
  Overview: 'A tour of the dashboard', Activity: 'Activity and history', Changes: 'Changes',
  Integrate: 'The Integrate page and the two briefs', Providers: 'Connections and keys', Routing: 'Routing and protocol defaults',
  Agents: 'A tour of the dashboard', Compression: 'A tour of the dashboard', Tasks: 'A tour of the dashboard', Settings: 'A tour of the dashboard',
};

/** A small info button that opens the manual at section. */
export function HelpLink({section}: {section: string}) {
  const open = useContext(HelpContext);
  return <button type="button" className="help-link" onClick={() => open(section)} title={'In the manual: ' + section} aria-label={'Open the manual at: ' + section}>
    <Info size={14} strokeWidth={1.75} aria-hidden/>
  </button>;
}
