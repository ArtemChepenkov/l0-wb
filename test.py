from flask import Flask, render_template, send_file


app = Flask(__name__)

@app.route("/")
def func():
	return send_file("./web/index.html")

app.run(host="0.0.0.0", port="8080")
