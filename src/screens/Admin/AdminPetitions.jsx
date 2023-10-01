import React from "react";
import { auth, db } from "../../firebase";

//styles
import "../../styles/scss/Admin/petition.scss";

function AdminPetitions() {
  const [petition, setPetition] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [goal , setGoal] = React.useState(0);
  const [petitions, setPetitions] = React.useState([]);


  const addPetition = (e) => {
    e.preventDefault();
    db.collection("petitions").add({
      title: e.target.petitionTitle.value,
      goal: e.target.petitionGoal.value,
      signatures: 0,
      users: [],
    }).then(() => {
      setGoal(0);
      setPetition("");
    })
  }

  React.useEffect(() => {
    setLoading(true);
    db.collection("petitions").onSnapshot((snapshot) => {
      setPetitions(snapshot.docs.map((doc) => ({
        id: doc.id,
        title: doc.data().title,
        goal: doc.data().goal,
        signatures: doc.data().signatures,
        users: doc.data().users,
      })))
    })
    setLoading(false);
  }
  , [])


  const handleDelete = (id) => {
    db.collection("petitions").doc(id).delete();
  }

  

  return(
    <div className="petitionsContainer">
      <form className="petitionsForm"
        onSubmit={addPetition}
      >
        <label for="petitionTitle">Petition Title</label>
        <input type="text" id="petitionTitle" name="petitionTitle" placeholder="Petition Title"/>
        <label for="petitionGoal">Petition Goal</label>
        <input type="number" id="petitionGoal" name="petitionGoal" placeholder="Petition Goal"/>
        <input type="submit" value="Submit" className="petitionSubmit"/>
      </form>
      <h1>Current Petitions</h1>
      <div className="petitionsList">
        {petitions.map((petition) => (
          <div className="petitionCard">
            <h3>{petition.title}</h3>
            <p>Signatures: {petition.signatures}</p>
            <p>Goal: {petition.goal}</p>
            <button onClick={() => handleDelete(petition.id)}>Delete</button>

              <h4>Users:</h4>
              {petition.users.map((user) => (
                <p>{user.fullName}</p>
              ))}
          </div>
        ))}
      </div>
    </div>
  )
}

export default AdminPetitions;
