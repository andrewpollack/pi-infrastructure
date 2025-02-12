export interface Meal {
    Day: number;
    Meal: string;
    URL: string | null;
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
    nextMonthResponse: MonthResponse;
}
