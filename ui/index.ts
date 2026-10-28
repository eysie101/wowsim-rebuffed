import './shared/bootstrap_overrides';

import * as Popper from '@popperjs/core';
import * as bootstrap from 'bootstrap';
import { Chart, registerables } from 'chart.js';

import changelog from '../CHANGELOG.md?raw';

Chart.register(...registerables);
Chart.defaults.color = 'white';

declare global {
	interface Window {
		Popper: any;
		bootstrap: any;
	}
}

window.Popper = Popper;
window.bootstrap = bootstrap;

// Force scroll to top when refreshing
if (history.scrollRestoration) {
	history.scrollRestoration = 'manual';
} else {
	window.onbeforeunload = function () {
		window.scrollTo(0, 0);
	};
}

function docReady(fn: any) {
	// see if DOM is already available
	if (document.readyState === 'complete' || document.readyState === 'interactive') {
		// call on next available tick
		setTimeout(fn, 1);
	} else {
		document.addEventListener('DOMContentLoaded', fn);
	}
}

function renderChangelog(text: string) {
	const entries = document.getElementById('changelog-entries');
	if (!entries) {
		return;
	}
	let list: HTMLUListElement | null = null;
	let item: HTMLLIElement | null = null;
	for (const line of text.split(/\r?\n/)) {
		if (line.startsWith('## ')) {
			const time = document.createElement('h4');
			time.className = 'changelog-time';
			time.textContent = line.slice(3).trim();
			list = document.createElement('ul');
			item = null;
			entries.append(time, list);
		} else if (line.startsWith('- ') && list) {
			item = document.createElement('li');
			item.textContent = line.slice(2).trim();
			list.append(item);
		} else if (/^\s+\S/.test(line) && item) {
			item.textContent += ' ' + line.trim();
		}
	}
}

docReady(function () {
	renderChangelog(changelog);
	document.body.classList.add('ready');
});
