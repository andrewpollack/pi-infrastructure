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

export interface CalendarResponse {
	currMonthResponse: MonthResponse;
}

export enum StatusType {
	LOADING = 'LOADING',
	ERROR = 'ERROR',
	SUCCESS = 'SUCCESS',
	UNKNOWN = 'UNKNOWN'
}
