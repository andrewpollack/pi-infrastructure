import random
import os
from datetime import datetime
import calendar
import json
import typing as T

import boto3

from http.server import BaseHTTPRequestHandler, HTTPServer


BUCKET_NAME = os.environ["BUCKET_NAME"]
OBJ_NAME = "recipes.json"


class Recipes:
    def __init__(self) -> None:
        s3 = boto3.resource("s3")
        content_object = s3.Object(BUCKET_NAME, OBJ_NAME)
        file_content = content_object.get()["Body"].read().decode("utf-8")
        self.recipes = json.loads(file_content)

        self.scratch_space = []

    def shuffle(self) -> None:
        random.shuffle(self.recipes)
        for i in range(len(self.recipes)):
            random.shuffle(self.recipes[i]["items"])

    def reset(self) -> None:
        self.recipes = self.scratch_space
        self.scratch_space = []

    def get(self, i) -> None:
        return self.recipes[i]["items"]

    def append(self, item, remove_from_temp: False) -> None:
        if remove_from_temp:
            self.scratch_space.pop()
        self.recipes.append(item)

    def pop(self) -> T.Dict[T.AnyStr, T.Union[T.AnyStr, T.List[T.AnyStr]]]:
        group = self.recipes.pop()
        self.scratch_space.append(group)
        return group

    def __len__(self) -> int:
        return len(self.recipes)


class SimpleHTTPRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()
        self.wfile.write(
            self.response_string.encode()
        )


def generateMeals(current_month, current_year):
    recipes = Recipes()

    random.seed(current_month + current_year)
    meal_calendar = calendar.monthcalendar(current_year, current_month)

    recipes.shuffle()

    for curr_week in meal_calendar:
        for curr_day_index in range(len(curr_week)):
            if curr_week[curr_day_index] == 0:
                curr_week[curr_day_index] = "NONE"
                continue

            if curr_day_index == calendar.FRIDAY + 1:
                curr_week[curr_day_index] = "Out"
                continue

            if curr_day_index == calendar.THURSDAY + 1:
                curr_week[curr_day_index] = "Leftovers"
                continue

            if len(recipes) < 1:
                recipes.reset()
                recipes.shuffle()

            selected_group = recipes.pop()
            selected_item = selected_group["items"].pop()
            selected_item = (
                f'<a href="{selected_item["url"]}">{selected_item["name"]}</a>'
                if selected_item["url"] != ""
                else selected_item["name"]
            )

            curr_week[curr_day_index] = selected_item

            add_group_back = False
            if selected_group["category"] == "Multiple okay":
                flip = random.randint(0, 100)
                if flip <= 40 and len(recipes) >= 3:
                    # Yay! Include special again.
                    add_group_back = True

            if len(selected_group) < 1:
                continue

            if add_group_back:
                recipes.append(selected_group, True)
                recipes.shuffle()

    return meal_calendar


def makeTableEntity(items, month, year):
    html_output = f"""
<h1>{month} {year}</h1>
<table>
    <tr>
        <th>Sunday</th>
        <th>Monday</th>
        <th>Tuesday</th>
        <th>Wednesday</th>
        <th>Thursday</th>
        <th>Friday</th>
        <th>Saturday</th>
    </tr>
"""
    for sublist in items:
        html_output += "  <tr>\n"
        for item in sublist:
            html_output += f"    <td>{item}</td>\n"
        html_output += "  </tr>\n"

    html_output += "</table>"
    return html_output


def main():
    calendar.setfirstweekday(calendar.SUNDAY)
    current_month = datetime.now().month
    current_month_name = calendar.month_name[current_month]
    current_year = datetime.now().year

    if current_month == 12:
        next_month = 1
        next_year = current_year + 1
    else:
        next_month = current_month + 1
        next_year = current_year

    next_month_name = calendar.month_name[next_month]

    this_month_meals = generateMeals(current_month, current_year)
    next_month_meals = generateMeals(next_month, next_year)

    server_class = HTTPServer
    handler_class = SimpleHTTPRequestHandler
    this_month_meals_table = makeTableEntity(
        this_month_meals, current_month_name, current_year
    )
    next_month_meals_table = makeTableEntity(
        next_month_meals, next_month_name, next_year
    )

    handler_class.response_string = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Meals</title>
    <style>
        table, th, td {{
            border: 1px solid black;
        }}
    </style>
</head>
<body>


{this_month_meals_table}

{next_month_meals_table}

</body>
</html>
"""
    port = int(os.environ["SERVE_PORT"])
    server_address = ("", port)
    httpd = server_class(server_address, handler_class)
    print(f"Starting httpd server on port {port}...")
    httpd.serve_forever()


if __name__ == "__main__":
    main()
