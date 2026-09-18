const motion = document.querySelector("#motion");

let paused = window.matchMedia(
  "(prefers-reduced-motion: reduce)"
).matches;

function updateMotion() {
  document.body.classList.toggle("motion-off", paused);
  motion.setAttribute("aria-pressed", String(paused));

  motion.innerHTML = paused
    ? '<span class="pause-icon">▷</span> Riprendi il volo'
    : '<span class="pause-icon">Ⅱ</span> Ferma il volo';
}

updateMotion();

motion.addEventListener("click", () => {
  paused = !paused;
  updateMotion();
});

const works = {
  city: {
    title: "Neon District",
    category: "001 / ENVIRONMENT",
    description:
      "Una città senza orizzonte. Tra nebbia, cemento e insegne " +
      "luminose, ogni finestra custodisce una storia ancora da raccontare.",
    alt: "Concept art di una metropoli cyberpunk notturna"
  },

  airship: {
    title: "Silent Voyager",
    category: "002 / VEHICLE DESIGN",
    description:
      "Sopra il rumore, un viaggio lento. Un dirigibile attraversa " +
      "il cielo notturno, portando con sé le luci di una città lontana.",
    alt: "Concept art di un dirigibile futuristico"
  }
};

const dialog = document.querySelector("#art-dialog");

document.querySelectorAll("[data-art]").forEach((button) => {
  button.addEventListener("click", () => {
    const key = button.dataset.art;
    const work = works[key];

    document.querySelector("#detail-title").textContent = work.title;
    document.querySelector("#detail-category").textContent = work.category;
    document.querySelector("#detail-description").textContent =
      work.description;

    const image = document.querySelector("#detail-image");
    image.src = `assets/${key}.png`;
    image.alt = work.alt;

    dialog.setAttribute("aria-labelledby", "detail-title");
    dialog.showModal();
  });
});

document.querySelector(".close").addEventListener("click", () => {
  dialog.close();
});

dialog.addEventListener("click", (event) => {
  if (event.target !== dialog) return;

  const bounds = dialog.getBoundingClientRect();

  const clickedOutside =
    event.clientX < bounds.left ||
    event.clientX > bounds.right ||
    event.clientY < bounds.top ||
    event.clientY > bounds.bottom;

  if (clickedOutside) {
    dialog.close();
  }
});

// Gestione invio trasmissione al server Go
const contactForm = document.querySelector("#contact-form");
if (contactForm) {
  const messageInput = document.querySelector("#contact-message");

  contactForm.addEventListener("submit", async (e) => {
    e.preventDefault();

    const submitBtn = document.querySelector("#contact-submit");
    const btnText = submitBtn.querySelector(".btn-text");
    const btnIcon = submitBtn.querySelector(".btn-icon");
    const btnLoading = submitBtn.querySelector(".btn-loading");
    const feedback = document.querySelector("#contact-feedback");

    const payload = {
      name: document.querySelector("#contact-name").value.trim(),
      email: document.querySelector("#contact-email").value.trim(),
      message: messageInput.value.trim()
    };

    // Stato di caricamento
    submitBtn.disabled = true;
    if (btnText) btnText.style.display = "none";
    if (btnIcon) btnIcon.style.display = "none";
    if (btnLoading) btnLoading.style.display = "inline";
    feedback.className = "form-feedback";
    feedback.textContent = "";

    try {
      const response = await fetch("/api/contact", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(payload)
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Impossibile inviare il messaggio.");
      }

      feedback.className = "form-feedback success";
      feedback.textContent = "Messaggio inviato con successo. Grazie per aver scritto!";
      contactForm.reset();
    } catch (err) {
      feedback.className = "form-feedback error";
      feedback.textContent = "Errore durante l'invio: " + err.message;
    } finally {
      submitBtn.disabled = false;
      if (btnText) btnText.style.display = "inline";
      if (btnIcon) btnIcon.style.display = "inline";
      if (btnLoading) btnLoading.style.display = "none";
    }
  });
}