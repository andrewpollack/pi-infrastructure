### garage-go

I (too often) leave for work, round the corner, and immediately second guess whether I closed the garage door. So, I fixed this concern.

This Garage door monitoring system uses a Raspberry Pi 2 W hooked up with a magnetic door sensor. The magnetic door sensor detects when the garage door is in the "closed" position. Metrics are exposed via
a Prometheus exporter, which are scraped by a Prometheus deployment I'm using across these projects. Finally, a dashboard is available to view the yes/no question through Grafana.

![image](https://github.com/user-attachments/assets/0fc0b76a-0609-489c-9992-567ffb8a7b0f)

Hardware portion was super helpful to follow the book <ins>[Automate Your Home Using Go](https://pragprog.com/titles/gohome/automate-your-home-using-go/)</ins>
