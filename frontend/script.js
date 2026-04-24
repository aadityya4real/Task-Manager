
const BASE_URL = "https://task-manager-78yn.onrender.com";

// 🔐 SIGNUP
function signup() {
  const username = document.getElementById("username").value;
  const password = document.getElementById("password").value;

  fetch(`${BASE_URL}/signup`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      username,
      password
    })
  })
  .then(res => res.json())
  .then(data => {
    alert("Signup successful");
    console.log(data);
  })
  .catch(err => console.error("Signup error:", err));
}

// 🔐 LOGIN
function login() {
  fetch(`${BASE_URL}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      username: document.getElementById("username").value,
      password: document.getElementById("password").value
    })
  })
  .then(res => res.json())
  .then(data => {
    console.log("LOGIN RESPONSE:", data);

    if (!data.token) {
      alert("Login failed");
      return;
    }

    localStorage.setItem("token", data.token);
    document.getElementById("auth").style.display = "none";
    document.getElementById("app").style.display = "block";
    loadTasks();
  })
  .catch(err => console.error("LOGIN ERROR:", err));
}

// 📥 LOAD TASKS
function loadTasks() {
  fetch(BASE_URL + "/tasks", {
    headers: {
      "Authorization": "Bearer " + localStorage.getItem("token")
    }
  })
  .then(res => res.json())
  .then(result => {

    const tasks = result.data;
    const list = document.getElementById("taskList");
    list.innerHTML = "";

    tasks.forEach(task => {
      const li = document.createElement("li");

      li.innerHTML = `
        <input type="checkbox" 
          ${task.done ? "checked" : ""} 
          onchange="toggleTask(${task.id}, '${task.title}', this.checked)" />

        <span class="${task.done ? 'done' : ''}">
          ${task.title}
        </span>

        <button onclick="deleteTask(${task.id})">❌</button>
      `;

      list.appendChild(li);
    });
  });
}
// ➕ ADD TASK
function addTask() {
  const title = document.getElementById("taskInput").value;

  fetch(`${BASE_URL}/tasks`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + localStorage.getItem("token")
    },
    body: JSON.stringify({ title })
  })
  .then(res => res.json())
  .then(() => loadTasks())
  .catch(err => console.error("ADD ERROR:", err));
}

// ❌ DELETE TASK
function deleteTask(id) {
  fetch(`${BASE_URL}/tasks?id=${id}`, {
    method: "DELETE",
    headers: {
      "Authorization": "Bearer " + localStorage.getItem("token")
    }
  })
  .then(() => loadTasks());
}

// 🔄 TOGGLE TASK
function toggleTask(id, title, done) {
  fetch(`${BASE_URL}/tasks?id=${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + localStorage.getItem("token")
    },
    body: JSON.stringify({
      title: title,
      done: done
    })
  }).then(() => loadTasks());
}

// ✏️ EDIT TASK
function editTask(id, oldTitle) {
  const newTitle = prompt("Edit task:", oldTitle);
  if (!newTitle) return;

  fetch(`${BASE_URL}/tasks?id=${id}`, {
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

// 🚪 LOGOUT
function logout() {
  localStorage.removeItem("token");
  location.reload();
}

// 🔄 AUTO LOGIN
window.onload = function () {
  const token = localStorage.getItem("token");
  if (token) {
    document.getElementById("auth").style.display = "none";
    document.getElementById("app").style.display = "block";
    loadTasks();
  }
};