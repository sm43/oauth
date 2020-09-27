import React from "react";
import Button from '@material-ui/core/Button';
import GitHubLogin from "react-github-login";
import { API_URL, GH_CLIENT_ID } from "./config.js";

const onSuccess = (response) => {
  const authorizeCode = response.code.toString();
  fetch(`${API_URL}/oauth/redirect?code=${authorizeCode}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
  })
    .then(function (response) {
      if (!response.ok) {
        alert("Login failed. Try again !!");
        throw new Error("Login failed - ", response);
      }
      return response.json();
    })
    .then((response) => {
      localStorage.setItem("token", response.token.toString());
      var x = document.getElementById("logoutDiv");
      x.style.display = "block";
      x = document.getElementById("loginDiv");
      x.style.display = "none";
    })
    .catch((error) => {
      console.error("Error:", error);
    });
};
const onFailure = (error) => {
  alert("Login failed !! Try again !!");
  console.log(error);
};


const logout = () => {
  var x = document.getElementById("loginDiv");
  x.style.display = "block";
  x = document.getElementById("logoutDiv");
  x.style.display = "none";
  x = document.getElementById("det");
  x.style.display = "none";
  localStorage.setItem("token", "");
};


export const Auth = () => {
  return (
    <div>
      <br></br>
      <div id="loginDiv">
        <GitHubLogin
          clientId={GH_CLIENT_ID}
          redirectUri=""
          onSuccess={onSuccess}
          onFailure={onFailure}
          id="login"
        />
      </div>
      &emsp;
      <br></br>
      <div id="logoutDiv" style={{ display: "none" }}>
        <button id="logout" name="logout" type="submit" onClick={logout}>
          Logout
          </button>
        <br></br>

        <br></br>

        <br></br>
        User Login Successfull..
      </div>
    </div>

  );
};


const show = () => {
  if (localStorage.getItem("token") !== "") {
    fetch(`${API_URL}/details`, {
      method: "GET",
      headers: {
        'Authorization': 'Bearer ' + localStorage.getItem("token"),
        "Content-Type": "application/json",
      },
    })
      .then(function (response) {
        if (!response.ok) {
          alert("failed to get user!!");
          throw new Error("failed to get user - ", response);
        }
        return response.json();
      })
      .then((response) => {
        var x = document.getElementById("det");
        x.style.display = "block";
        document.getElementById("name").innerHTML = response.name.toString()
        document.getElementById("gh-id").innerHTML = response.githubID.toString()
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  } else {
    alert("Please login first...!!");
  }

}
export const Detail = () => {
  return (
    <div>
      <br></br>
      <Button variant="contained" onClick={show}>Show User Details</Button>
      <div id="det" style={{ display: "none" }}>
        <label>Name: </label><p id="name"></p>
        <label>GiHub ID: </label><p id="gh-id"></p>
      </div>
    </div>
  );
}
