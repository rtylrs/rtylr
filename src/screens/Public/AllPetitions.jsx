import React, { useEffect, useState } from "react";
import { db } from "../../firebase";
import PetitionModal from "../../components/PetitionModal";

function AllPetitions() {
  const [isPetitionModalOpen, setIsPetitionModalOpen] = useState(false);
  const [petitions, setPetitions] = useState([]);
  const [petitonId, setPetitionId] = useState("");
  const [products, setProducts] = useState([]);

  useEffect(() => {
    const petitionsUnsubscribe = db
      .collection("petitions")
      .onSnapshot((petitionsSnapshot) => {
        const petitionsData = petitionsSnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        setPetitions(petitionsData);
      });

    return () => {
      petitionsUnsubscribe();
    };
  }, []);

  return (
    <div>
      <PetitionModal
        isOpen={isPetitionModalOpen}
        onClose={() => {
          setIsPetitionModalOpen(false);
        }}
        petitionId={petitonId}
      />

      <section className="petition-section">
        <h2>Latest Petitions</h2>
        <div className="petition-list">
          {petitions.map((petition, index) => (
            <div className="petition-item" key={index}>
              <h3>{petition.title}</h3>
              <div className="progress-container">
                <div className="progress-section-one">
                  <div className="signature-progress">
                    {petition.signatures} / {petition.goal}
                  </div>
                  <div className="progress-bar">
                    <div
                      className="progress"
                      style={{
                        width: `${
                          petition.signatures >= petition.goal
                            ? "100%"
                            : `${Math.min(
                                (petition.signatures / petition.goal) * 100,
                                100
                              )}%`
                        }`,
                      }}
                    ></div>
                  </div>
                </div>
                <div className="progress-section-two">
                  <button
                    className="sign-button"
                    onClick={() => {
                      setPetitionId(petition.id);
                      setIsPetitionModalOpen(true);
                    }}
                  >
                    Sign Now
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}

export default AllPetitions;
