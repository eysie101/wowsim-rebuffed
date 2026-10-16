import { TypedEvent } from '../typed_event.js';

import { Input, InputConfig } from './input.js';

export interface RangePickerConfig<ModObject> extends InputConfig<ModObject, number> {
	min: number;
	max: number;
	// Text shown next to the slider for a value.
	formatValue: (value: number) => string;
}

// A slider over whole numbers, with the current value shown next to it.
export class RangePicker<ModObject> extends Input<ModObject, number> {
	private readonly rangeElem: HTMLInputElement;
	private readonly valueElem: HTMLSpanElement;
	private readonly formatValue: (value: number) => string;

	constructor(parent: HTMLElement, modObject: ModObject, config: RangePickerConfig<ModObject>) {
		super(parent, 'range-picker-root', modObject, config);
		this.formatValue = config.formatValue;

		this.rangeElem = document.createElement('input');
		this.rangeElem.type = 'range';
		this.rangeElem.classList.add('range-picker-input', 'form-range');
		this.rangeElem.min = String(config.min);
		this.rangeElem.max = String(config.max);
		this.rangeElem.step = '1';
		this.rootElem.appendChild(this.rangeElem);

		this.valueElem = document.createElement('span');
		this.valueElem.classList.add('range-picker-value');
		this.rootElem.appendChild(this.valueElem);

		this.init();

		this.rangeElem.addEventListener('input', () => {
			this.valueElem.textContent = this.formatValue(this.getInputValue());
		});
		this.rangeElem.addEventListener('change', () => {
			this.inputChanged(TypedEvent.nextEventID());
		});
	}

	getInputElem(): HTMLElement {
		return this.rangeElem;
	}

	getInputValue(): number {
		return parseInt(this.rangeElem.value);
	}

	setInputValue(newValue: number) {
		this.rangeElem.value = String(newValue);
		this.valueElem.textContent = this.formatValue(newValue);
	}
}
