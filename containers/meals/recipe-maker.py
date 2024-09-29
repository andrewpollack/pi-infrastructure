import base64
import calendar
import json
import os
import pickle
import random
import time
from datetime import date, datetime
from email.mime.text import MIMEText
from http.server import BaseHTTPRequestHandler, HTTPServer
import typing as T

import boto3
from googleapiclient.discovery import build
# from google.auth.transport.requests import Request
# from google.oauth2.credentials import Credentials
# from google_auth_oauthlib.flow import InstalledAppFlow


# If modifying these SCOPES, delete the file token.pickle.
SCOPES = ["https://www.googleapis.com/auth/gmail.send"]

BUCKET_NAME = os.environ["BUCKET_NAME"]
OBJ_NAME = "recipes.json"

AISLES_ORDERING = [
    "Cheese & Bakery",
    "18 & 19 (Alcohol, Butter, Cheese)",
    "16 & 17 (Freezer)",
    "10-15 (No Food Items)",
    "6-9 (Bevs & Snacks)",
    "3-5 (Breakfast & Baking)",
    "1 & 2 (Pasta, Global, Canned)",
    "Produce",
    "Meat & Yogurt",
]


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
        self.wfile.write(self.response_string.encode())


def generateMeals(current_month, current_year):
    recipes = Recipes()

    random.seed(current_month + current_year)
    meal_calendar = calendar.monthcalendar(current_year, current_month)

    recipes.shuffle()

    for curr_week in meal_calendar:
        for curr_day_index in range(len(curr_week)):
            if curr_week[curr_day_index] == 0:
                curr_week[curr_day_index] = {"name": "NONE"}
                continue

            if curr_day_index == calendar.FRIDAY + 1:
                curr_week[curr_day_index] = {"name": "Out"}
                continue

            if curr_day_index == calendar.THURSDAY + 1:
                curr_week[curr_day_index] = {"name": "Leftovers"}
                continue

            if len(recipes) < 1:
                recipes.reset()
                recipes.shuffle()

            selected_group = recipes.pop()
            selected_item = selected_group["items"].pop()
            # Potential overrides here...

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
        for selected_item in sublist:
            item = (
                f'<a href="{selected_item["url"]}">{selected_item["name"]}</a>'
                if "url" in selected_item.keys() != ""
                else selected_item["name"]
            )
            html_output += f"    <td>{item}</td>\n"
        html_output += "  </tr>\n"

    html_output += "</table>"
    return html_output


def runWebServer(this_month_meals_table, next_month_meals_table):
    server_class = HTTPServer
    handler_class = SimpleHTTPRequestHandler

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


def authenticate_gmail():
    creds = None
    if os.path.exists(os.environ["PICKLE_PATH"]):
        with open(os.environ["PICKLE_PATH"], "rb") as token:
            token_content = token.read()
            decoded_content = base64.b64decode(token_content)
            creds = pickle.loads(decoded_content)

    # NOTE: Keeping this here for when we need to regen creds.
    # # If no valid credentials available, let the user log in
    # if not creds or not creds.valid:
    #     if creds and creds.expired and creds.refresh_token:
    #         creds.refresh(Request())
    #     else:
    #         flow = InstalledAppFlow.from_client_secrets_file("credentials.json", SCOPES)
    #         creds = flow.run_local_server(port=0)

    #     # Save the credentials for the next run
    #     with open("token.pickle", "wb") as token:
    #         pickle.dump(creds, token)

    return build("gmail", "v1", credentials=creds)


def create_message(sender, to, subject, message_text):
    message = MIMEText(message_text)
    message["to"] = to
    message["from"] = sender
    message["subject"] = subject
    raw = base64.urlsafe_b64encode(message.as_bytes()).decode()
    return {"raw": raw}


def send_message(service, sender, to, subject, message_text):
    message = create_message(sender, to, subject, message_text)
    max_attempts = 5

    for attempt in range(1, max_attempts + 1):
        try:
            message = (
                service.users().messages().send(userId=sender, body=message).execute()
            )
            print(f'Message Id: {message["id"]}')
            return message
        except Exception as e:
            print(f"Attempt {attempt} failed: {e}")
            if attempt < max_attempts:
                print(f"Retrying in 5 seconds...")
                time.sleep(5)  # Wait 3 seconds before retrying
            else:
                print("Operation failed after 5 attempts.")
    return None


def generateEmailView(this_month_meals, next_month_meals, curr_week_index):
    this_weeks_meals = this_month_meals[curr_week_index]
    potential_next_month_meals = next_month_meals[0]

    # Need to take into account bleedover across months
    curr_list = [
        this_weeks_meals[i]
        if this_weeks_meals[i]["name"] != "NONE"
        else potential_next_month_meals[i]
        for i in range(7)
    ]

    meals_per_day = [
        "* Sunday: ",
        "* Monday: ",
        "* Tuesday: ",
        "* Wednesday: ",
        "* Thursday: ",
        "* Friday: ",
        "* Saturday: ",
    ]
    for i in range(len(curr_list)):
        meals_per_day[i] += curr_list[i]["name"]

    grocery_collection = {}
    for aisle in AISLES_ORDERING:
        grocery_collection[aisle] = {}

    for curr_item in curr_list:
        if "ingredients" not in curr_item:
            continue

        for ingredient in curr_item["ingredients"]:
            aisle, item, quantity, unit = (
                ingredient["aisle"],
                ingredient["item"],
                ingredient["quantity"],
                ingredient["unit"],
            )
            k = f"{item}__{unit}"
            if k not in grocery_collection[aisle]:
                grocery_collection[aisle][k] = (0, [])
            quant, related_meals = grocery_collection[aisle][k]
            quant += quantity
            related_meals += [curr_item["name"]]
            grocery_collection[aisle][k] = (quant, related_meals)

    formatted_collection = "Meals:\n"
    formatted_collection += "\n".join(meals_per_day)
    formatted_collection += "\n\n"

    for aisle in AISLES_ORDERING:
        items = []
        for k in sorted(grocery_collection[aisle].keys()):
            v = grocery_collection[aisle][k]
            quant, related_meals = v
            item, unit = k.split("__")
            items.append(f"* {quant} {unit}: {item} ({', '.join(related_meals)})")
        formatted_collection += f"{aisle}:\n"
        formatted_collection += "\n".join(items)
        if len(items) == 0:
            formatted_collection += "NONE"
        formatted_collection += "\n\n"

    return formatted_collection


def main():
    calendar.setfirstweekday(calendar.SUNDAY)
    current_month = datetime.now().month
    current_month_name = calendar.month_name[current_month]
    current_year = datetime.now().year

    # Finding which week it currently is
    first_day = date.today().replace(day=1)
    dom = date.today().day
    adjusted_dom = (
        dom + first_day.weekday()
    )  # Adjust day to account for start of the week
    curr_week_index = (adjusted_dom - 1) // 7

    if current_month == 12:
        next_month = 1
        next_year = current_year + 1
    else:
        next_month = current_month + 1
        next_year = current_year

    next_month_name = calendar.month_name[next_month]

    this_month_meals = generateMeals(current_month, current_year)
    next_month_meals = generateMeals(next_month, next_year)

    this_month_meals_table = makeTableEntity(
        this_month_meals, current_month_name, current_year
    )
    next_month_meals_table = makeTableEntity(
        next_month_meals, next_month_name, next_year
    )

    if "JUST_EMAIL" in os.environ:
        email_text = generateEmailView(
            this_month_meals, next_month_meals, curr_week_index
        )
        service = authenticate_gmail()
        sender = os.environ["SENDER_EMAIL"]
        to = os.environ["RECEIVER_EMAILS"]
        subject = f"Groceries for {current_month_name} Week {curr_week_index}"
        message_text = email_text

        if "DRY_RUN" not in os.environ:
            send_message(service, sender, to, subject, message_text)
        else:
            print(f"I would've sent the following from {sender} to {to}...\n\n")
            print(message_text)
    else:
        runWebServer(this_month_meals_table, next_month_meals_table)


if __name__ == "__main__":
    main()
