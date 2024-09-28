# recipe-maker.py

Quick and dirty script for randoomly selecting and rendering a months worth of recipes.
Deployed on a container so I can bring up the website locally from my phone or laptop
(whichever is more convenient at the time).

Recipes are stored in a JSON file in a private repository shared with my partner.
This repo has CD setup to push the latest state to S3 (for fun!), which is then pulled
by this script to make sure recipes are populated.
