function redirectToEditPage(name) {
    if (name === undefined) {
        window.location.href = "./edit";
    } else {
        window.location.href = "./edit?name=" + name;
    }
}

function fetchData() {
    fetch('/api/recasters')
        .then(response => response.json())
        .then(data => {
            const recasterContainer = document.getElementById('recaster-index');
            console.log(data)
            data.forEach(recaster => {
                const listItem = document.createElement('li');

                const label = document.createElement('label');
                label.appendChild(document.createTextNode(recaster.Name));
                listItem.appendChild(label)

                const button = document.createElement('button');
                button.onclick = function() {
                    redirectToEditPage(recaster.Name)
                }
                button.textContent = 'Edit'
                listItem.appendChild(button)

                recasterContainer.appendChild(listItem)
            });
        })
}

window.onload = fetchData;
