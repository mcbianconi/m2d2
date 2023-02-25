![tests](https://github.com/mcbianconi/m2d2/actions/workflows/test.yaml/badge.svg)


# Objetivo
:construction: WIP

Substituir referências a diagramas [D2](https://d2lang.com/) presentes em arquivos markdown
por sua versão renderizada.


# Exemplo
Em um arquivo markdown, o seguinte diagrama

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

seria compilado e substituído por uma tag do tipo: `![label](link-para-o-diagrama.svg)`

