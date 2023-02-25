![tests](https://github.com/mcbianconi/m2d2/actions/workflows/test.yaml/badge.svg)


# Objetivo
:construction: WIP

Substituir referências a diagramas [D2](https://d2lang.com/) presentes em arquivos markdown
por sua versão renderizada.


# Exemplo

Em um arquivo markdown

```d2
direction: left

a <-> b

b -> d: foi {
    style {
        multiple: true
        animated: true
    }
}

e -> f <- b

b.shape: cloud
b.style.multiple: true


b <- a: selva {
    style.animated: true
}

```

O bloco de código acima, que define um diagrama, seria substituído por uma tag do tipo: `![label](link-para-o-diagrama.svg)` se começasse com ` ```d2`
