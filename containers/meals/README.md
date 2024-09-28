# recipe-maker.py

![Screenshot from 2024-09-28 15-54-16](https://github.com/user-attachments/assets/92b2241f-ee41-4184-aa17-0ba6494cf091)

Quick and dirty script for randoomly selecting and rendering a months worth of recipes.
Deployed as a Deployment+NodePort Service so I can bring up the website locally from my
phone or laptop (whichever is more convenient at the time).

Recipes are stored in a JSON file in a private repository shared with my partner.
This repo has CD setup to push the JSON file's latest state to S3, which is then pulled
by this script to make sure recipes are populated. While we could just pull from the GitHub
repo itself, CD in this way is far more fun!

![Screenshot from 2024-09-28 13-27-38](https://github.com/user-attachments/assets/4b9abc7b-37e7-4730-8e1a-121b2c9d3536)
