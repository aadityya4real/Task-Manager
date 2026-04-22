const API = "http://localhost:8080";

function signup() {
  fetch(API + "/signup", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      username: document.getElementById("username").value,
      password: document.getElementById("password").value
    })
  })
  .then(res => res.json())
  .then(data => alert("Signup successful"));
}

function login() {
  fetch(API + "/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      username: document.getElementById("username").value,
      password: document.getElementById("password").value
    })
  })
  .then(res => res.json())
  .then(data => {
    localStorage.setItem("token", data.token);
    document.getElementById("auth").style.display = "none";
    document.getElementById("app").style.display = "block";
    loadTasks();
  });
}

function loadTasks() {
  fetch(API + "/tasks", {
    headers: {
      "Authorization": "Bearer " + localStorage.getItem("token")
    }
  })
  .then(res => res.json())
  .then(result => {
    console.log("GET response:", result); // 🔥 debug

    const tasks = result.data; // 🔥 IMPORTANT

    const list = document.getElementById("taskList");
    list.innerHTML = "";

    tasks.forEach(task => {
      const li = document.createElement("li");

      li.innerHTML = `
        ${task.title}
        <button onclick="deleteTask(${task.id})">Delete</button>
      `;

      list.appendChild(li);
    });
  })
  .catch(err => console.error("LOAD ERROR:", err));
}

function addTask() {
  const title = document.getElementById("taskInput").value;

  fetch(API + "/tasks", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + localStorage.getItem("token")
    },
    body: JSON.stringify({ title })
  })
  .then(res => res.json())
  .then(data => {
    console.log("POST response:", data); // 🔥 debug
    loadTasks();
  })
  .catch(err => console.error("ADD ERROR:", err));
}

function deleteTask(id) {
  fetch(API + "/tasks?id=" + id, {
    method: "DELETE",
    headers: {
      "Authorization": "Bearer " + localStorage.getItem("token")
    }
  })
  .then(() => loadTasks());
}

function logout() {
  localStorage.removeItem("token");
  location.reload();
}
window.onload = function () {
  const token = localStorage.getItem("token");
  if (token) {
    document.getElementById("auth").style.display = "none";
    document.getElementById("app").style.display = "block";
    loadTasks();
  }
};
function toggleTask(id, title, done) {
  fetch(API + "/tasks?id=" + id, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + localStorage.getItem("token")
    },
    body: JSON.stringify({
      title: title,   // ✅ KEEP ORIGINAL TITLE
      done: done
    })
  }).then(() => loadTasks());
}
function editTask(id, oldTitle) {
  const newTitle = prompt("Edit task:", oldTitle);
  if (!newTitle) return;

  fetch(API + "/tasks?id=" + id, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + localStorage.getItem("token")
    },
    body: JSON.stringify({
      title: newTitle,
      done: false
    })
  }).then(() => loadTasks());
}