export interface Meal {
	Day: number;
	Meal: string;
	URL: string | null;
	Enabled: boolean | null;
}

export interface MealsResponse {
	allMeals: Meal[];
}

export interface MonthResponse {
	Year: number;
	Month: string;
	MealsEachWeek: Meal[][];
}

export interface ExtraItem {
	Name: string;
	Aisle: string;
	ID: number;
	Enabled: boolean;
}

export type ExtraItemUpdate = {
	Action: 'Add' | 'Update' | 'Delete';
	Old: ExtraItem | null;
	New: ExtraItem | null;
};

export interface ExtraItemsResponse {
	allItems: ExtraItem[];
}

// Type for the page data including extra items and aisles
export interface AislesResponse {
	extraItems: ExtraItem[];
	aisles: string[];
}

export interface CalendarResponse {
	currMonthResponse: MonthResponse;
}

export enum StatusType {
	LOADING = 'LOADING',
	ERROR = 'ERROR',
	SUCCESS = 'SUCCESS',
	UNKNOWN = 'UNKNOWN'
}

export interface EmailsResponse {
	emails: string[];
}
