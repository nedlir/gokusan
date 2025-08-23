import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Button } from "../../UI/Button";
import { registerUser } from "../../http/api";

function Register() {
  const [formData, setFormData] = useState({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });

  const { mutate, error, isPending } = useMutation({
    mutationFn: registerUser,
    onSuccess: (data) => {
      if (data.success) {
        alert(data.message || "Registration successful!");
      }
    },
  });

  if (isPending) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <div>Register</div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (formData.password !== formData.confirmPassword) {
            alert("Passwords do not match");
            return;
          }
          mutate(formData);
        }}
      >
        <div>
          <label htmlFor="username">Username:</label>
          <input
            type="text"
            id="username"
            value={formData.username}
            onChange={(e) =>
              setFormData((prev) => ({ ...prev, username: e.target.value }))
            }
            autoComplete="username"
            required
          />
        </div>
        <div>
          <label htmlFor="email">Email:</label>
          <input
            type="email"
            id="email"
            value={formData.email}
            onChange={(e) =>
              setFormData((prev) => ({ ...prev, email: e.target.value }))
            }
            autoComplete="email"
            required
          />
        </div>
        <div>
          <label htmlFor="password">Password:</label>
          <input
            type="password"
            id="password"
            value={formData.password}
            onChange={(e) =>
              setFormData((prev) => ({ ...prev, password: e.target.value }))
            }
            autoComplete="new-password"
            required
          />
        </div>
        <div>
          <label htmlFor="confirmPassword">Confirm Password:</label>
          <input
            type="password"
            id="confirmPassword"
            value={formData.confirmPassword}
            onChange={(e) =>
              setFormData((prev) => ({
                ...prev,
                confirmPassword: e.target.value,
              }))
            }
            autoComplete="new-password"
            required
          />
        </div>
        <Button type="submit">Register</Button>
      </form>
    </div>
  );
}

export default Register;
